# SIMD (Single Instruction, Multiple Data) Specification

## Overview

The Nora Programming Language supports high-performance mathematical and data-parallel operations through native SIMD integrations. Instead of relying purely on auto-vectorization from the backend C compiler, Nora allows developers to explicitly define hardware-backed vector types that enforce CPU SIMD execution (such as SSE, AVX, AVX2, and AVX-512).

## Motivation

For tight numerical workloads (game engines, scientific computing, spectral normalizations), manual memory allocations, fiber checkpoints, and scalar mathematical operations cause severe performance bottlenecks. By introducing native SIMD support:
1. Developers have granular control over vector layouts.
2. The compiler emits direct CPU vector extensions in the generated C11 backend (`__attribute__((vector_size(N)))`).
3. The Nora inliner explicitly strips fiber checkpoints (cooperative yields) inside heavily optimized vector equations.

## Syntax

SIMD support is exposed through the `[vector_size(N)]` attribute on tuple-struct declarations. `N` defines the total byte-size of the vector. 
For example, a struct of four 64-bit floats (`f64`) requires 32 bytes of alignment (4 * 8 = 32):

```nora
[vector_size(32)]
pub type Vec4d = struct {
    x: f64
    y: f64
    z: f64
    w: f64
}
```

## Semantics and Type Rules

### Vector Size Validation
The compiler ensures that `vector_size(N)` strictly matches the memory layout of the primitive fields within the struct. If `N` does not match the byte-size of the fields combined, the compiler will panic.

### Native Operators
Structs annotated with `[vector_size(N)]` automatically gain support for native binary arithmetic operators without requiring manual method implementations. 
- **Addition (`+`)**: Maps to SIMD packed addition.
- **Subtraction (`-`)**: Maps to SIMD packed subtraction.
- **Multiplication (`*`)**: Maps to SIMD packed multiplication.
- **Division (`/`)**: Maps to SIMD packed division.

```nora
var a = simd.Set4(1.0, 2.0, 3.0, 4.0)
var b = simd.Set1(5.0)

// Native overloaded SIMD math
var c = a + b 
var d = a * b
```

### Equality and Memory Layout
Vectors can be evaluated for equality `a == b`. Under the hood, Nora transpiles vector equality into `memcmp` comparisons within the C compiler, as vector structs bypass traditional C scalar equivalence.

## Compiler Optimizations & Inlining

When defining mathematical formulas over SIMD types, developers **must** use the `[inline]` attribute. 
Because Nora is a cooperatively scheduled language, every non-inlined function forces an atomic `NR_COOPERATIVE_YIELD_CHECKPOINT()` macro inside the loop body, destroying the CPU pipeline and preventing Clang/GCC from unrolling and vectorizing the block.

**Bad Approach (High Overhead):**
```nora
fn calculate(a: Vec4d, b: Vec4d) Vec4d {
    // Hidden atomic load caused by cooperative fiber checkpoint
    return a + b
}
```

**Correct Approach (Hardware Native):**
```nora
[inline]
fn calculate(a: Vec4d, b: Vec4d) Vec4d {
    // Inliner merges this directly into the parent loop. No checkpoints!
    return a + b
}
```

## Conditional Compilation 

SIMD features can be hardware-gated using conditional compilation attributes:
```nora
[cfg(target_feature = "avx")]
pub fn fast_path() { ... }
```
When running or building with `--target-feature avx`, Nora explicitly passes `-mavx` to the C compiler, ensuring the backend enables the correct vector CPU instruction sets.

## Examples
See `examples/spectralnorm/simd` for a fully benchmarked AVX-capable implementation of the Spectral Norm algorithm.
