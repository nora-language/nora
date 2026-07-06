# Investigation: Struct Literal HIR Bypass Bug

**Status**: Root Cause Found, Partially Fixed
**Date**: July 6, 2026
**Component**: Codegen, HIR, AST

## Problem
When compiling `nora_wgpu`'s `triangle` example, the C-compiler reports errors for unexpected type names (e.g. `sys_WGPUTextureFormat _tmp = sys_WGPUTextureFormat.BGRA8Unorm`) and undefined function calls (e.g., `ffi_BorrowToRaw_5905e0f6(NULL, color_target)`). 

While the primitive enum `[repr("i32")]` bug was fixed previously in the HIR Lowering layer, these errors still occur specifically inside **struct initialization blocks**.

## Root Cause
Nora's compiler utilizes two parallel code generation paths:
1. **High-Level IR (HIR) Lowering** (`pkg/hir/lower.go` -> `pkg/codegen/hir_codegen.go`)
2. **Legacy AST stringifier** (`pkg/codegen/expressions.go` - `exprToString`)

Currently, `StructLiteral` nodes (e.g., `alloc sys.WGPUSurfaceConfiguration { format: sys.WGPUTextureFormat.BGRA8Unorm }`) **skip HIR lowering entirely**. Because they skip the HIR layer, the `exprToString` generation kicks in and evaluates every struct field directly.

This bypass causes several critical failures:
1. **Implicit Moves**: The Lease Solver detects enum accesses (value types) inside struct literals as "implicit moves". The AST stringifier's `genSelectorExpression` attempts to execute this move by doing `memset(&sys_WGPUTextureFormat.BGRA8Unorm, 0, ...)` which is invalid for constants/constructors. (Note: *This has been fixed*).
2. **Generic Function Instantiation**: Generic functions like `ffi.BorrowToRaw[sys.WGPUFragmentState]` called inside struct literals are hashed by `analyzer.go` and assigned a specialized name (`ffi_BorrowToRaw_5905e0f6`), but because HIR lowering is skipped, the Code Generator NEVER receives the `requestGenericFunction` trigger to actually emit the C-code for this function.
3. **Pointer Type Erasure**: Type erasure logic (generating `_ptr` suffix for pointers) also fails to trigger correctly when falling back to the AST generator.

## Reproduction
Any generic function call or newly added HIR feature applied inside a struct literal will bypass HIR lowering and fail.

```nora
var raw = ffi.BorrowToRaw[sys.WGPUFragmentState](#fragment) // Works (HIR handled)

var cfg = alloc sys.WGPUFragmentState {
    targets: ffi.BorrowToRaw[sys.WGPUColorTargetState](#color_target) // Fails (AST stringifier handled, generic body never emitted)
}
```

## Fix (Next Steps)
There are two potential fixes:
1. **Architectural Fix**: Add support for `*ast.StructLiteral` directly into `pkg/hir/lower.go` so that struct literals utilize the HIR code generation pipeline properly.
2. **Patch Fix**: Ensure that `genCallExpression` inside `pkg/codegen/expressions.go` triggers generic function body emission when it detects a generic call that hasn't been emitted yet.

We should pursue the architectural fix of adding `StructLiteral` to the HIR lowerer, as the dual codegen paths will continue to cause disjointed feature support.
