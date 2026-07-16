# Investigation Report: SIMD Native Types & Inliner Bug

## Status
Resolved

## Problem
When attempting to optimize the `spectralnorm` benchmark using SIMD (`[native("__m256d")]`), the Nora compiler failed to generate valid C code, resulting in multiple `clang` compilation errors. Specifically, three distinct issues occurred:
1. **Move Semantics on Native Types**: Passing native SIMD vectors (e.g., `Vec4d`) to functions threw `use of moved value` errors because the semantic analyzer treated them as owned structs instead of copyable primitives.
2. **C Type Redefinition & Initialization**: The codegen emitted `typedef struct simd_Vec4d simd_Vec4d`, which conflicted with the native `__m256d` definition in `nora_simd.h`. Furthermore, it attempted to initialize native vectors using aggregate struct initializers (`{.v0 = ..., .v1 = ...}`), which C prohibits for `__m256d`.
3. **Inliner Void Return Bug**: When `[inline]` functions that return `void` (e.g., `simd.Store`) were processed by the HIR Inliner (`pkg/hir/optimize/inliner.go`), the compiler generated temporary variables to hold their "return" value. This resulted in `void _inline_var_XYZ;` declarations and invalid statements in the generated C code, causing the C compiler to fail with `variable has incomplete type 'void'`.

## Reproduction
1. Define a struct with a native attribute:
```nora
[native("__m256d")]
pub type Vec4d = struct {
    v0: f64
    v1: f64
    v2: f64
    v3: f64
}
```
2. Create an inline function returning void:
```nora
[inline]
pub fn Store(data: ptr, vec: Vec4d) {
    nr_simd_store(data, vec)
}
```
3. Call the inline function. The compiler generates invalid C code:
```c
void _inline_var_223; // Error: incomplete type 'void'
_inline_var_223;      // Error: undeclared identifier
```

## Root Cause
- **Frontend**: The semantic analyzer (`pkg/types/types.go`) did not recognize `NativeType` structs as simple, copyable values (`IsOwnedType` returned true for all structs).
- **Codegen**: The C generator blindly emitted forward declarations (`pkg/codegen/prototypes.go`) and field-by-field equality checks (`pkg/codegen/generator.go`) for all structs, disregarding whether they mapped to native C types.
- **Inliner (`[inline]`)**: The inlining pass (`pkg/hir/optimize/inliner.go`) lacked a check for `void` return types. It universally allocated a return variable (`retVar := inl.cloner.NextTemp()`) and passed it downstream, forcing the C generator to emit a statement for a `void` temporary variable.

## Fix
1. **Semantic Analyzer**: Modified `IsOwnedType` in `pkg/types/types.go` to return `false` if `st.NativeType != ""`.
2. **Codegen Adjustments**: 
   - Skipped emitting forward declarations (`emitCombinedTypeDefs`) for `NativeType` structs.
   - Updated equality operator generation to use `memcmp` for `NativeType` structs instead of field-by-field comparison.
   - Added `simd.Set4` and replaced aggregate initializers in `spectralnorm.nr`.
3. **Inliner Patch**: Updated `pkg/hir/optimize/inliner.go` to skip allocating a return variable (`Alloca`) and returning a mock `hir.Expression` if the inlined function's return type is `void`:
```go
if i.Type != nil && !types.Equals(i.Type, types.Void) {
    return &hir.Expression{Expr: retVar, Type: i.Type}
}
return nil
```

## Validation
After applying the fixes, compiling `examples/spectralnorm/simd/spectralnorm.nr` with `--release --cflags "-mavx"` successfully generated `build/release/spectralnorm.exe`. 
Execution time dropped from **~1 minute** to **~1.2 seconds**, matching the C reference implementation's performance.
