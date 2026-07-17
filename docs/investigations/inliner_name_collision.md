# Investigation: Inliner Name Collision Bug

## Status
Unresolved (Pending Compiler Fix)

## Problem
Adding `pub fn Sqrt` to the `std/simd` library caused `spectralnorm.nr` to fail compilation with a type mismatch error (`passing 'double' to parameter of incompatible type 'simd_Vec4d'`). The compiler was incorrectly substituting calls to `math.Sqrt(f64)` with calls to the C built-in for `simd.Sqrt(Vec4d)` (i.e. `nr_simd_sqrt`).

## Reproduction
1. Ensure both `math.Sqrt` and an `[inline]` function named `simd.Sqrt` exist.
2. In a program that imports both `math` and `simd`, call `math.Sqrt(x)` where `x` is a `f64`.
3. Compile the program. The C Codegen will output a call to the inline body of `simd.Sqrt` instead of `math.Sqrt`.

## Root Cause
The Nora compiler's HIR inlining pass (`pkg/hir/optimize/inliner.go`) contains a critical name collision bug:
1. The inliner populates a dictionary of inlineable functions (`inliner.hirFuncs`). 
2. Instead of keying this dictionary with the fully qualified mangled name (`pkg_math_Sqrt`), it explicitly overrides the key with the short un-mangled source name (`FuncSymbol.Name`, which is just `"Sqrt"`).
3. Because both `math.Sqrt` and `simd.Sqrt` share the same short name `"Sqrt"`, `simd.Sqrt` overwrites `math.Sqrt` in the global dictionary.
4. When the inliner encounters a call to `math.Sqrt`, it queries the dictionary for `"Sqrt"`, receives the HIR representation for `simd.Sqrt`, and incorrectly inlines it.

## Fix
Currently unresolved in the compiler codebase due to architectural constraints on how `FuncName` and `FuncSymbol` are assigned during AST-to-HIR lowering. 

**Workaround**:
Avoid using identical short names for `[inline]` functions in the standard library. For example, `simd.Sqrt` was commented out or must be renamed to `simd.VecSqrt` to prevent it from colliding with `math.Sqrt`.

## Validation
By commenting out `simd.Sqrt`, the name collision was eliminated. `spectralnorm.nr` successfully compiled and execution time returned to its expected baseline of ~2.0 seconds, confirming that the performance regression caused by an incomplete compiler fix (which broke all inlining) was resolved.
