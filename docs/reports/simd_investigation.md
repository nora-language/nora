# Investigation Report: Nora Standard Library `std/simd`

## 1. Current State & The "Unsupported CPU" Problem
Currently, Nora's `std/simd` package is hardcoded to the **x86_64 AVX** instruction set. 

**Evidence from the codebase:**
- `nora.yaml` injects `cflags: ["-mavx"]`.
- `nora_simd.h` uses `#pragma GCC target("avx")` and `#include <x86intrin.h>`.
- The `simd.nr` primitives map explicitly to `__m256d` (256-bit AVX vectors) via the `[native("__m256d")]` attribute.
- Functions like `nr_simd_approx_recip` rely heavily on specific x86 intrinsic sequences (`_mm256_cvtpd_ps`, `_mm_rcp_ps`, etc.).

### What happens if the CPU is not supported?
1. **On non-x86 architectures (e.g., ARM/Apple Silicon):** The compilation will **fail entirely** during the C11 compilation phase because `<x86intrin.h>` is missing and GCC/Clang will not recognize the `avx` target.
2. **On older x86 CPUs (without AVX):** If the program is compiled on a machine with AVX but executed on an older machine, the OS will trigger a **`SIGILL` (Illegal Instruction) crash** the moment an AVX register is used.

---

## 2. Best Solutions for the Nora Compiler
To adhere to Nora's core philosophy—*Correctness, language consistency, and compiler maintainability are far more important than implementation speed*—we need a SIMD solution that is **portable by default**, but allows hardware-specific optimizations when required.

Here is the recommended multi-tiered architecture for Nora's compiler:

### A. The Ideal Solution: GCC/Clang Generic Vectors
Instead of mapping Nora types directly to hardware registers (`__m256d`), we should leverage the fact that Nora transpiles to C11 using GCC/Clang. Both compilers support **Generic Vector Extensions**.

**Proposed Nora Syntax:**
```nora
// Inform the compiler this is a generic SIMD vector of 4 f64 elements
[vector_size(4)] 
pub type Vec4d = f64 
```

**Proposed C11 Codegen:**
The Nora compiler would lower this attribute to:
```c
typedef double simd_Vec4d __attribute__((vector_size(32)));
```

**Why is this the best approach?**
- **Zero Fallbacks Needed:** When using `__attribute__((vector_size(X)))`, standard C operators (`+`, `-`, `*`, `/`) work out of the box. 
- **Auto-Adaptation:** If compiled for ARM, Clang emits NEON instructions. If compiled for modern x86, it emits AVX2. If the CPU has no SIMD, the C compiler automatically lowers it to a scalar loop.
- **Maintainability:** `simd.nr` operations like `Add(a, b)` can just be written in pure Nora (`return a + b`), dropping the need for hundreds of `extern fn` bindings in `nora_simd.h`.

### B. Hardware-Specific Intrinsics Subpackages
Generic vectors cover 80% of use cases (math, logic). However, algorithms like Spectral Norm rely on specialized operations (like `nr_simd_approx_recip`, swizzling, and blending) that don't map to generic vectors.

Nora should introduce architecture-specific subpackages for these:
- `std/simd/x86`
- `std/simd/arm`

### C. Conditional Compilation `[cfg(...)]`
To tie it all together, Nora's parser and semantic analyzer need to be upgraded to support conditional compilation attributes (similar to Rust). 

This allows `std/simd` to provide optimized routines when the CPU supports it, while providing a pure scalar Nora implementation as a fallback.

```nora
[cfg(target_feature = "avx")]
pub fn ApproxRecip(z: Vec4d) Vec4d {
    // Call into x86 intrinsics
    return x86.ApproxRecip(z)
}

[cfg(not(target_feature = "avx"))]
pub fn ApproxRecip(z: Vec4d) Vec4d {
    // Pure scalar math fallback (Newton-Raphson approximation)
    // Ensures the code runs ANYWHERE.
}
```

## 3. Implementation Roadmap
If we choose to proceed with upgrading Nora's SIMD capabilities, we should follow this plan:

1. **Phase 1: Compiler Attributes:** Add `[cfg(target_arch = "...")]` and `[cfg(target_feature = "...")]` to `pkg/parser` and evaluate them in `pkg/semantic`.
2. **Phase 2: Generic Vectors:** Introduce the `[vector_size(N)]` attribute and update `pkg/codegen` to emit GCC vector attributes.
3. **Phase 3: Refactor `std/simd`:** Rewrite the standard library to use the new generic vectors for arithmetic, moving x86-specific intrinsics behind `[cfg]` gates with scalar fallbacks.
