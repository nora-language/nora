# Defer Statement Direct Codegen Implementation Plan

## Title
Defer Statement Support in Direct C Code Generation (`pkg/codegen`)

## Status
Completed

## Metadata
- **Author:** Antigravity
- **Date:** 2026-07-14
- **Components:** `pkg/codegen` (`generator.go`, `statements.go`)

## Goal
To fully implement the `defer` keyword (`ast.DeferStatement`) in Nora's direct C code generator (`pkg/codegen`). While `defer` is already fully lowered inside the High-level Intermediate Representation (`pkg/hir`) pipeline, the direct C code generator (`pkg/codegen`) currently stubs `genDeferStatement` (`pkg/codegen/statements.go:1040`) with a C comment (`/* defer not implemented in direct codegen */`). Implementing `defer` in `pkg/codegen` guarantees that deferred function calls execute reliably in Last-In, First-Out (LIFO) order when exiting a function scope or returning early via `return`.

## Affected Compiler Components
- `pkg/codegen/generator.go`: Add `ActiveDefers []ast.Expression` to the `Generator` struct. Manage saving, clearing, and restoring `ActiveDefers` across function and lambda boundaries (`genFunction`). Emit remaining active deferred expressions at function termination points right before RAII scope drops.
- `pkg/codegen/statements.go`:
  - Update `genDeferStatement()` (`case *ast.DeferStatement`) to push `s.Call` onto `g.ActiveDefers`.
  - Update `genReturnStatement()` (`case *ast.ReturnStatement`) to emit all deferred expressions from `g.ActiveDefers` in reverse order (`i := len(g.ActiveDefers)-1; i >= 0; i--`) immediately before `g.emitReturnDrops(drops)`.

## Implementation Checklist
### 1. Update `Generator` State in `pkg/codegen/generator.go`
- [x] Add `ActiveDefers []ast.Expression` field to the `Generator` struct definition (`generator.go`).
- [x] In `genFunction()`, save `prevDefers := g.ActiveDefers`, reset `g.ActiveDefers = nil`, and restore `defer func() { g.ActiveDefers = prevDefers }()`.
- [x] At the end of `genFunction()` (before emitting parameter/receiver drops and closing the function block), check if the block did not end in a trailing `ReturnStatement`. If not, emit all calls in `g.ActiveDefers` in LIFO order (`for i := len(g.ActiveDefers)-1; i >= 0; i-- { g.genExpression(...) }`).

### 2. Implement `genDeferStatement()` in `pkg/codegen/statements.go`
- [x] Replace `/* defer not implemented in direct codegen */` in `genDeferStatement()` with `g.ActiveDefers = append(g.ActiveDefers, s.Call)`.

### 3. Update `genReturnStatement()` in `pkg/codegen/statements.go`
- [x] In `genReturnStatement()`, after evaluating the return value expression (`_ret = ...`), check `if len(g.ActiveDefers) > 0`.
- [x] If `len(g.ActiveDefers) > 0`, emit each deferred call in LIFO order right before `g.emitReturnDrops(drops)` (for both `void` and non-`void` returns).

### 4. Verification & Regression Testing
- [x] Create a positive integration test in `pkg/cmd/test/defer_codegen_test/defer_codegen_test.nr` verifying LIFO execution and proper behavior on early return vs fallthrough.
- [x] Create a negative integration test in `pkg/cmd/test/defer_codegen_test/fail_defer_non_call.nr` verifying diagnostic errors on invalid syntax.

## Test Plan
1. **Positive Test (`pkg/cmd/test/defer_codegen_test/defer_codegen_test.nr`)**:
   - Define a function with multiple `defer` calls (`defer io.PrintLn("1")`, `defer io.PrintLn("2")`). Verify they run in LIFO order (`2` then `1`).
   - Define a function with early return (`if x > 0 { return }`). Verify `defer` executes right before the early return.
2. **Negative Test (`pkg/cmd/test/defer_codegen_test/fail_defer_non_call.nr`)**:
   - Verify that non-call expressions (e.g., `defer x + 1`) trigger parser/semantic errors.

## Risks
- Interaction between `defer` calls and RAII variable drops (`PreDrops` / `ReturnDrops`): Variables referenced by deferred calls must remain valid when the deferred call executes. Nora's Topological Lease Solver (`solver.go:826`) already anchors all variables used inside `defer` calls (`usages`) to `AnchorEndOfFunction`, guaranteeing that RAII destructors run *after* the deferred function calls complete.

## Completion Criteria
- [x] `genDeferStatement()` pushes deferred expressions to `ActiveDefers`.
- [x] Deferred expressions execute in exact LIFO order before `return` and at the end of functions.
- [x] Positive and negative tests created under `pkg/cmd/test/defer_codegen_test/`.
