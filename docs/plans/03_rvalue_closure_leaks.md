# Implementation Plan: Fix Inline Closure Environment Leaks (R-Value Drops)

## Status
Pending

## Goal
Fix the 8-byte memory leak that occurs when passing closures as inline arguments (e.g. `DoMap(fn(...) { ... })`). In Nora, closure environments are allocated on the heap using `nr_malloc` and compile to fat pointers (`nr_closure_t`). Currently, the Topological Lease Solver only schedules drops for tracked named variables (`ast.VarStatement`, etc.). R-values, such as anonymous inline closures passed as arguments, are ignored, leaking their dynamically allocated environment structures.

## Affected Compiler Components
* `pkg/topology/solver.go`

## Proposed Changes

### 1. `pkg/topology/solver.go`
*   Modify `Solve()` or `analyzeBlock()` to track temporary r-values generated within expressions, specifically focusing on `ast.LambdaExpression` which allocates memory.
*   Introduce an `AnonymousDrops` tracking mechanism. When a statement is processed, any nested `ast.LambdaExpression` that is not immediately bound to a named variable (e.g., used directly as a function argument) must have a drop scheduled for it at the end of the statement or block.
*   Ensure that closures are properly consumed or freed. If the closure is passed as an owned type (e.g. `@fn()`), the receiving function consumes it and drops it. If it is passed as a read-only lease (e.g. `#fn()`), as is common in higher-order functions like `DoMap`, the caller retains ownership and MUST inject an `nr_free` or a closure drop method for the r-value at the end of the caller's statement.

## Implementation Checklist
- [ ] Add tracking of inline `ast.LambdaExpression` nodes during `analyzeBlock` expression traversal.
- [ ] Schedule drops for inline closures at the end of the current expression statement.
- [ ] Update `hir_codegen.go` or `generator.go` to generate the correct C drop logic (`nr_free(closure.env)`) for temporary `nr_closure_t` instances.
- [ ] Validate that the memory leak in `pkg/cmd/test/pass_closure_capture_generic` is resolved and drops to 0 bytes.

## Test Plan
- Re-run `TestCompilerWithTestFolder/pass_closure_capture_generic`.
- Ensure no memory leaks are reported by `nr_mem_report`.
- Create a new test case ensuring nested inline closures (closures inside closures passed as arguments) do not leak.
