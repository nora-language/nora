# SIMD Standard Library Roadmap

This roadmap outlines the long-term plan for evolving the Nora `std/simd` module from its current foundational state into a complete, production-ready vectorization library comparable to Rust's `std::simd` or C++'s intrinsic libraries.

## Phase 1: Foundation (Completed ✅)
The core architecture has been established, relying on generic compiler vector extensions (`[vector_size(N)]`) and conditional compilation (`[cfg("target_feature=avx")]`).

- [x] **Core Types**: `Vec4d` (f64x4), `Vec8f` (f32x8), `Vec4f` (f32x4), `Vec8i` (i32x8), `Vec4i` (i32x4).
- [x] **Memory Ops**: Basic contiguous `Load` and `Store`.
- [x] **Basic Arithmetic**: Native operator overloads for `+`, `-`, `*`, `/`.
- [x] **Basic Math & Logic**: `Min`, `Max`, `Sqrt`, `ApproxRecip`, `And`, `Or`, `Xor`, `AndNot`.
- [x] **Testing**: End-to-end integration tests verifying correct C-lowering and AVX instruction emission.

## Phase 2: Expanded Data Types & Widths
Broaden the library to support all hardware-supported primitives and wider AVX-512 registers.

- [x] **Small Integers (8-bit)**: `Vec16i8`, `Vec32i8`, `Vec16u8`, `Vec32u8`
- [x] **Small Integers (16-bit)**: `Vec8i16`, `Vec16i16`, `Vec8u16`, `Vec16u16`
- [x] **Unsigned 32-bit Integers**: `Vec4u32`, `Vec8u32` (Missing from Phase 1)
- [x] **Large Integers (64-bit)**: `Vec2i64`, `Vec4i64`, `Vec2u64`, `Vec4u64`
- [x] **AVX-512 Vectors (512-bit)**: `Vec16f`, `Vec8d`, `Vec16i32`, `Vec16u32`, `Vec8i64`, `Vec8u64`
    - Require `[cfg("target_feature=avx512f")]`
    - Add matching typedefs in `nora_simd.h` when `__AVX512F__` is available.

## Phase 3: Comparisons & Masking (Completed ✅)
Enable branching-free conditional logic within SIMD execution.

- [x] **Comparison Operators**: Functions for `CmpEq`, `CmpNeq`, `CmpLt`, `CmpGt`.
- [x] **Mask Types**: Introduction of SIMD mask types (e.g., `Mask8`, `Mask4`) to represent vectors of booleans.
- [x] **Masked Execution**: Implement masked variants of operations (e.g., `AddMasked(a, b, mask)` which only adds elements where the mask is true).
- [x] **Dynamic Blending**: Functions to dynamically blend two vectors based on a runtime mask.

## Phase 4: Advanced Memory Operations (Completed ✅)
Support complex data layout ingestions.

- [x] **Gather**: Load elements from non-contiguous memory locations using a vector of indices.
- [x] **Scatter**: Write elements to non-contiguous memory locations.
- [x] **Alignment Specifications**: `LoadAligned` and `StoreAligned` for optimized memory access when memory is known to be 16/32-byte aligned.

## Phase 5: Shuffles, Swizzles & Permutations (Completed ✅)
Allow developers to arbitrarily rearrange data within vectors.

- [x] **Intra-lane Swizzling**: Rearrange elements within 128-bit lanes.
- [x] **Cross-lane Permutations**: Full AVX/AVX2 permute capabilities to cross 128-bit lane boundaries.
- [x] **Unpacking/Interleaving**: Standard functions to interleave the high or low halves of two vectors (useful for matrix transpositions).

## Phase 6: Conversions & Horizontal Reductions
Facilitate type shifting and collapsing vectors into scalar values.

- [x] **Type Casting**: Fast conversions between types (e.g., converting `Vec8i` to `Vec8f`).
- [x] **Width Casting**: Extending/truncating between widths (e.g., `Vec4f` to `Vec4d`).
- [x] **Horizontal Reductions**: 
  - [x] `Sum()`: Returns the sum of all elements as a scalar (implemented for `Vec8d` as `VecReduceAdd8d`).
  - [x] `Product()`: Returns the product of all elements.
  - [x] `MaxElement()` / `MinElement()`: Finds the largest/smallest scalar in the vector.

## Phase 7: Complex Mathematical Approximations
Implement standard transcendental math functions using high-performance polynomial approximations (Taylor/Minimax series).

- [x] **Trigonometry**: `Sin`, `Cos`.
- [x] **Logarithmic**: `Log`, `Log2`.
- [x] **Exponential**: `Exp`.
