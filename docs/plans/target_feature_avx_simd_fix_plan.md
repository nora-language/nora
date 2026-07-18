# Target Feature (`--target-feature`) & SIMD Library Pragma Override Fix Plan

## Status
Proposed

## Metadata
- **Author:** Antigravity
- **Date:** 2026-07-18
- **Components:** `pkg/cmd/nora`, `std/simd`

## Goal
To ensure that the Nora compiler and C toolchain correctly honor `--target-feature` (`-mavx`, etc.) command-line flags when validating build cache entries and when compiling standard library SIMD helpers (`std/simd/nora_simd.h`). This eliminates illegal instruction (`SIGILL`) crashes when running compiled release binaries (`-O3`) on host architectures that support `AVX` but lack `AVX2` or `FMA`.

## Affected Compiler Components
- `pkg/cmd/nora/main.go`: Manages package build cache hashing and dependency caching (`catalog.Packages[pkg]` and `compileCSourceDependencyToCache`), where target features (`opts.Target.Features`) and custom CFlags (`activeConfig.CFlags`) must be included in the cache validation hash calculation.
- `std/simd/nora_simd.h`: Standard C header defining inline 128-bit, 256-bit, and 512-bit SIMD intrinsics and helper functions, which currently hardcodes blanket `#pragma GCC target("avx,avx2,fma")` and `#pragma clang attribute push (__attribute__((target("avx,avx2,fma"))), apply_to=function)`.

## Implementation Checklist
### 1. Update Build Cache Hashing in `pkg/cmd/nora/main.go`
- [ ] In `compileCSourceDependencyToCache()` (around line 1996), append `opts.Target.Features` (sorted) and `activeConfig.CFlags` to the hash input (`hashInput := ...`) before computing SHA-256 (`sha256.Sum256`).
- [ ] In package compilation cache validation (`catalog.Packages[pkg]` check around line 1830), ensure that `opts.Target.Features` and `activeConfig.CFlags` are incorporated into the package source hash / `currentHash` computation so that changing CLI `--target-feature` flags triggers a clean `Cache MISS` and recompilation.

### 2. Refactor Target Attributes in `std/simd/nora_simd.h`
- [ ] Replace the blanket `#pragma GCC target("avx,avx2,fma")` and `#pragma clang attribute push (__attribute__((target("avx,avx2,fma"))), apply_to=function)` at the top of `nora_simd.h` with architecture-safe conditional compilation or granular target attributes.
- [ ] Ensure that when only `__AVX__` is active (`--target-feature avx` / `-mavx`) without `__AVX2__`, functions like `nr_simd_set1(double val)` (`_mm256_set1_pd`) and `nr_simd_load` compile safely without generating `vbroadcastsd %xmm0, %ymm0` (`AVX2`).
- [ ] Guard `AVX512F` and higher-tier SIMD attributes cleanly without breaking existing AVX/AVX2 configurations or inliner assumptions.

### 3. Verification & Regression Testing
- [ ] Add a regression verification step confirming that compiling with `--target-feature avx` on an AVX-only host generates `.o` files without `vbroadcastsd` and executes `spectralnorm 5500` cleanly (`Exit code: 0`).
- [ ] Run `go test ./pkg/cmd/nora...` to verify that cache invalidation and SIMD compilation tests pass without regressions across standard builds.

## Test Plan
1. **Cache Invalidation Verification:** Build `spectralnorm.nr` with `--target-feature avx`, verify `Cache HIT` on consecutive builds, then switch to `--target-feature avx2` and confirm it produces a `Cache MISS` and re-invokes the C toolchain.
2. **SIMD Codegen Check:** Verify via `objdump -d` that when compiled with `-mavx -mno-avx2`, `nr_simd_set1` and related 256-bit vector helpers do not emit `AVX2` register-to-register broadcast opcodes.
3. **Integration Verification:** Run `./nora build --release --target-feature avx examples/spectralnorm/simd/spectralnorm.nr && ./build/release/spectralnorm 100` and confirm exact verification output match without `SIGILL`.

## Risks
- Removing top-level `#pragma GCC target("avx,avx2,fma")` means C compilers will rely on `-mavx` or `-mavx2` flags passed from `nora build` to unlock 256-bit SIMD intrinsics (`immintrin.h`). We mitigate this by ensuring `pkg/target/target.go` and `compileCToObject` consistently pass the appropriate `-m` flags based on `opts.Target.Features` across all POSIX platforms.

## Completion Criteria
- Implementation plan created and reviewed.
- Cache hashing fixes applied cleanly in `pkg/cmd/nora/main.go`.
- `std/simd/nora_simd.h` target pragmas updated to support AVX-only targets without SIGILL.
- All integration tests and verification checks pass.
