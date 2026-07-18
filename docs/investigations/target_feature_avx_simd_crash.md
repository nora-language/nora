# Investigation Report: `--target-feature` CLI Flag Overridden by Hardcoded SIMD Library Pragmas and Build Cache

## Status
Open

## Problem
When compiling and running `examples/spectralnorm/simd/spectralnorm.nr` with explicit AVX target features (`./nora build --release --target-feature avx examples/spectralnorm/simd/spectralnorm.nr`), the generated executable crashes immediately with `SIGILL` (`Signal 4: Illegal instruction`, Exit Code `132`) on machines that support `AVX` but not `AVX2`. Despite `--target-feature avx` (`-mavx`) being passed from the Nora CLI to the C compiler toolchain, AVX2 instructions (specifically `vbroadcastsd %xmm0, %ymm0`) are still emitted.

## Reproduction
1. On a host machine supporting up to `AVX` (without `AVX2` or `FMA` support, such as standard virtual machine instances or pre-Haswell x86_64 CPUs), verify SIMD flags:
```bash
grep -m 1 "flags" /proc/cpuinfo | grep -E -o 'avx|avx2|fma'
# Output shows only: avx
```
2. Build `spectralnorm.nr` specifying `--target-feature avx`:
```bash
./nora build --release --target-feature avx examples/spectralnorm/simd/spectralnorm.nr
```
3. Execute the resulting release binary:
```bash
./build/release/spectralnorm 5500
```
4. Observe immediate process termination with `SIGILL`:
```text
Command terminated by signal 4 (SIGILL / Illegal instruction)
Exit code: 132
```
5. Inspect the crash point using `gdb`:
```asm
=> 0x5555555681f0:  vbroadcastsd %xmm0,%ymm0  (# AVX2: VBROADCASTSD ymm, xmm/m64)
```

## Root Cause
Our investigation identified two independent issues across the compiler build driver (`pkg/cmd/nora/main.go`) and the standard SIMD library (`std/simd/nora_simd.h`):

1. **Build Cache Validation Omits Target Features (`pkg/cmd/nora/main.go`)**:
   The package compilation cache (`catalog.Packages[pkg]` check around line 1830) and the C dependency cache (`compileCSourceDependencyToCache` around line 1996) compute cache hash keys based on AST source hashes, optimization levels, and target platform (`TargetOS`/`TargetArch`). They do **not** incorporate `opts.Target.Features` (`--target-feature`) or `opts.CFlags` into the hash calculation. As a result, running `nora build --target-feature avx` after a previous build without the flag yields a `Cache HIT`, reusing stale object files (`cache_pkg_simd.o`) compiled with different SIMD assumptions without recompilation.

2. **Hardcoded Target Pragmas in Standard Library (`std/simd/nora_simd.h`)**:
   Even after manually purging the build cache (`rm -rf build/release/runtime_cache/*`) to force a clean rebuild with `-mavx`, Clang continues to emit `vbroadcastsd %xmm0, %ymm0`. Inspection of `std/simd/nora_simd.h` (lines 3, 5, and 1167) revealed hardcoded target attributes:
```c
#pragma GCC target("avx,avx2,fma")
#pragma clang attribute push (__attribute__((target("avx,avx2,fma"))), apply_to=function)
```
   These `#pragma` directives unconditionally force Clang and GCC to compile every SIMD inline helper function (`nr_simd_set1`, `nr_simd_load`, etc.) with `AVX2` and `FMA` instructions enabled inside the C AST. This overrides command-line target flags (`-mavx` or `-mno-avx2`). When `-O3` optimization runs, Clang vectorizes `_mm256_set1_pd(val)` (`nr_simd_set1`) directly into the AVX2 register-to-register broadcast instruction `vbroadcastsd %xmm0, %ymm0`.

## Fix
To resolve this issue cleanly across the compiler driver and standard runtime:

1. **Update Build Cache Hashing (`pkg/cmd/nora/main.go`)**:
   - Incorporate sorted `opts.Target.Features` and `activeConfig.CFlags` into the hash computation for both package object files and C dependencies (`catalog.Packages` and `compileCSourceDependencyToCache`). This ensures that altering `--target-feature` or `-cflags` correctly invalidates existing `.o` files and triggers a `Cache MISS`.

2. **Refactor `std/simd/nora_simd.h` Target Attributes**:
   - Remove the blanket `#pragma GCC target("avx,avx2,fma")` and `#pragma clang attribute push (__attribute__((target("avx,avx2,fma"))))` from the top-level header.
   - Rely on compiler command-line flags (`-mavx`, `-mavx2`, `-mfma`) to dictate available instruction sets, or guard pragmas and SIMD intrinsics with standard feature macros (`#if defined(__AVX2__)`).
   - Where `AVX2` specific register-to-register broadcast instructions are unavailable (`__AVX__` defined but `__AVX2__` absent), provide clean fallbacks (e.g., using `_mm256_setr_pd` / `_mm256_permute2f128_pd`) so that SIMD code compiles safely for pure `AVX` targets.

## Validation
- Verify `nora build --release --target-feature avx examples/spectralnorm/simd/spectralnorm.nr` on an AVX-only machine triggers recompilation when switching target features.
- Verify `objdump -d build/release/spectralnorm | grep vbroadcastsd` returns zero AVX2 register-to-register broadcast instructions when built with `-mavx -mno-avx2`.
- Verify `spectralnorm 5500` runs to completion and produces exact verified output (`1.274224153`) without triggering `SIGILL`.
