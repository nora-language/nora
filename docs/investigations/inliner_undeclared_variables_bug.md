# Compiler Investigation: Inliner Undeclared Variables Bug

**Status**: Root Cause Identified
**Problem**: When compiling with the `[inline]` attribute on functions that themselves contain function calls (nested inlining), the C compiler fails with `use of undeclared identifier '_inline_var_X'`.
**Reproduction**: 
1. Add `[inline]` to functions in `std/simd/simd.nr` (e.g. `simd.Set1`, `simd.Add`).
2. Add `[inline]` to `eval_A_vec` which calls these SIMD functions.
3. Compile a program that calls `eval_A_vec`.
4. Clang fails during C compilation due to disordered or missing `double _inline_var_X;` declarations.

## Root Cause
The Nora compiler operates with a Multi-Pass Architecture. The inlining pass (`pkg/hir/optimize/inliner.go`) manipulates the High-level Intermediate Representation (HIR).
1. **Global Temp Variables**: The `Cloner` struct shares a global `tempID`, generating globally unique variables like `_inline_var_1` for return variables and argument temps.
2. **Missing Symbol Association**: When `Alloca` instructions are created for these temps (`targetBlock.AddInst(&hir.Alloca{...})`), they are injected with `Symbol: nil`.
3. **Disordered AST Generation**: When nested functions are inlined, the inlined body (containing its own `Alloca` instructions) is cloned into the caller block. Because `processOperand` evaluates the RHS expressions *before* completing the parent assignment, and appends the inlined elements directly to `targetBlock`, multiple levels of cloning lead to `Alloca` instructions being appended out of dependency order relative to their actual assignments inside deep loop structures. 
4. **Locals Tracking Failure**: In C, variables must be declared before use. Because `hir_codegen.go` dynamically iterates the `HIRBlock.Elements` and emits declarations precisely where the `Alloca` appears in the AST, any AST reordering caused by nested cloning results in the `Assign` being emitted in C before the `Alloca`.

## Fix Plan
1. **Centralize Locals**: Instead of generating `hir.Alloca` instructions dynamically and letting them float around the block elements, the `Inliner` must explicitly append new temporary variables directly to the caller `hir.Function.Locals` map.
2. **Top-Level Declarations**: Modify `hir_codegen.go` to iterate over `hf.Locals` (the function's local variables map) and emit all C variable declarations at the absolute top of the C function body (C89 style), completely removing reliance on `hir.Alloca` position for declarations.
3. **Symbol Mapping**: Ensure `Cloner` creates and maintains proper `*semantic.Symbol` pointers for cloned variables to ensure drop flags and RAII behaviors are seamlessly inherited.

## Validation
Compile the SIMD version of `spectralnorm.nr` with all wrapper functions and `eval_A_vec` marked as `[inline]`. Verify that the emitted C code correctly hoists all `_inline_var_X` declarations to the top of the function and compiles without errors.
