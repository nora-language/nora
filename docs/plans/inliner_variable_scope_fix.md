# Implementation Plan: Inliner Variable Scope Bug

**Status**: Completed

## Goal
Fix the C compiler redefinition and undeclared identifier errors caused by the HIR inlining pass when compiling Nora functions marked with `[inline]`.

## Background
The `[inline]` attribute directs the Nora compiler to replace function calls with the actual body of the function. This is critical for high-performance abstractions (like SIMD wrapper functions). However, it introduces significant complexity because it requires rewriting the AST (Abstract Syntax Tree) to map the original function's arguments and local variables to newly generated unique temporary names (`_inline_var_X`) inside the caller's scope. 

This was a "big change" for two reasons:
1. **Variable Declaration Order**: C compilers are strict about variable scope. Because inlined elements were recursively injected into the AST, their `Alloca` (declaration) instructions were emitted out of order or deep inside loop blocks. If the code jumped or looped, C would throw an `undeclared identifier` error. We needed to fundamentally change `hir_codegen.go` to hoist all variable declarations to the absolute top of the generated C function.
2. **RAII Dependency Tracking**: Nora automatically inserts `Drop` instructions for memory safety. `Drop` instructions reference variables using `semantic.Symbol` pointers. When the inliner renamed variables to `_inline_var_X`, it was leaving the `Drop` instructions pointing to the original, un-renamed symbols. This caused the C generator to emit `nr_df_originalName` drop flags that didn't exist. We had to create a stateful symbol mapping within the `Cloner` struct.

## Implementation Checklist

- [x] Update `genFunction` in `pkg/codegen/hir_codegen.go` to recursively collect all `Alloca` instructions across the entire function block before emitting any other body instructions.
- [x] Modify `hir_codegen.go` to emit C variable declarations for all collected `Alloca` nodes immediately after the function prologue.
- [x] Manage `declaredVars` to prevent parameter/receiver variables and nested inlined variables from causing C "redefinition" errors for both values and their `nr_df_` drop flags.
- [x] Update `Cloner` struct in `pkg/hir/optimize/clone.go` to maintain a `symMap map[*semantic.Symbol]*semantic.Symbol`.
- [x] Update `CloneInstruction` for `*hir.Alloca` to generate a new `semantic.Symbol` for the inlined temporary and store it in `symMap`.
- [x] Update `CloneInstruction` for `*hir.Drop` and `*hir.VarOperand` to map back to the new `semantic.Symbol` using `symMap`.
- [x] Validate compilation of `examples/spectralnorm/simd/spectralnorm.nr` using `Nora build --release --cflags "-mavx"`.

## Testing
- Execute `spectralnorm.nr` with N=5000 to ensure performance scaling and correct execution.
