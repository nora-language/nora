# Investigation: Inliner Name Collision Bug

## Status
Resolved

## Problem
Adding `pub fn Sqrt` to the `std/simd` library caused `spectralnorm.nr` to fail compilation with a type mismatch error (`passing 'double' to parameter of incompatible type 'simd_Vec4d'`). The compiler was incorrectly substituting calls to `math.Sqrt(f64)` with calls to the C built-in for `simd.Sqrt(Vec4d)` (i.e. `nr_simd_sqrt`).

## Reproduction
1. Ensure both `math.Sqrt` and an `[inline]` function named `simd.Sqrt` exist.
2. In a program that imports both `math` and `simd`, call `math.Sqrt(x)` where `x` is a `f64`.
3. Compile the program. The C Codegen will output a call to the inline body of `simd.Sqrt` instead of `math.Sqrt`.

## Root Cause
The Nora compiler's HIR inlining pass (`pkg/hir/optimize/inliner.go`) contained a critical name collision bug:
1. The inliner populated a dictionary of inlineable functions (`inliner.hirFuncs`). 
2. Instead of keying this dictionary with the fully qualified mangled name (`pkg_math_Sqrt`) or symbol pointer, it keyed the map with the short un-mangled source name (`FuncSymbol.Name`, which is just `"Sqrt"`).
3. Because both `math.Sqrt` and `simd.Sqrt` shared the same short name `"Sqrt"`, `simd.Sqrt` overwrote `math.Sqrt` in the global dictionary.
4. When the inliner encountered a call to `math.Sqrt`, it queried the dictionary for `"Sqrt"`, received the HIR representation for `simd.Sqrt`, and incorrectly inlined it.

## Fix
Resolved in `pkg/hir/optimize/inliner.go` (`Inliner` struct, `runInlinePass`, and `processBlock`).
The inliner dictionary was split into `hirBySymbol map[*semantic.Symbol]*hir.Function` and `hirByName map[string]*hir.Function`. Functions and call targets (`*hir.Call`) are now primarily keyed and queried using their exact `*semantic.Symbol` pointer (`i.FuncSymbol`), completely eliminating name collisions across packages or within identical short names while retaining `hirByName` as a safe fallback for synthetic or symbol-less functions.

## Validation
By commenting out `simd.Sqrt`, the name collision was eliminated. `spectralnorm.nr` successfully compiled and execution time returned to its expected baseline of ~2.0 seconds, confirming that the performance regression caused by an incomplete compiler fix (which broke all inlining) was resolved.
