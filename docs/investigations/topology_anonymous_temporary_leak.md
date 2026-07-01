# Compiler Investigation: Topological Solver Fails to Drop Anonymous Temporaries

## Status
Unimplemented

## Problem
When a function returns an owned heap-allocated value (e.g., `@AABB[T]`) and the caller immediately chains a method or field access on the returned value without assigning it to a named variable (e.g., `func().Method()`), the Nora compiler's Topological Lease Solver fails to track the anonymous temporary variable. As a result, it fails to insert the necessary `Drop` instruction, causing a memory leak.

## Reproduction
To reproduce the issue, compile and run code that chains a method directly onto a function call returning an allocated object.

```nora
import "src/math/vector"

pub type Dummy = struct {
    val: @vector.Vector3[f64]
}

pub fn CreateDummy() @Dummy {
    var v = vector.NewVector3[f64](1.0, 1.0, 1.0)
    return alloc Dummy { val: v }
}

pub fn (d: #Dummy) Read() f64 {
    return d.val.x
}

pub fn TriggerLeak() {
    // BUG: The allocated Dummy is an anonymous temporary. 
    // The solver fails to drop the temporary after .Read() finishes.
    var val = CreateDummy().Read()
}
```

Running this code with memory debugging (`nora run --debug-memory`) will output a leak report showing that the `Dummy` struct (and its internal Vector) was never freed.

## Root Cause
The root cause resides in the Nora compiler's **Topological Lease Solver** (`pkg/topology`). 
The solver builds its dependency graph (`localLifecycles` and `allVisible`) primarily by tracking named variable births, assignments, and moves. When an expression contains an anonymous temporary (an r-value resulting from a function call), the semantic analyzer lowers it into an implicit temporary AST node. However, the Topological Solver does not properly register these implicit temporaries as independent lifecycles. Because the temporary lacks a registered dependency lifecycle, the solver's RAII drop insertion phase silently skips it, leaving the generated C code without a corresponding `nr_drop()` call at the end of the statement expression.

## Fix
*(This represents the required compiler-side fix)*
The AST lowering phase and semantic analyzer must assign explicit internal temporary names (e.g., `__tmp_0`) for all intermediate r-values that return owned types. The Topological Lease Solver (`pkg/topology/solver.go`) must be updated to track these internal temporaries exactly like normal named variables, ensuring that their births are recorded and `Drop` / `PreDrop` instructions are automatically inserted at the end of the full expression statement.

## Validation
* Create a positive integration test (`pkg/cmd/test/repro_anonymous_temporary_leak.nr`) containing the reproduction code.
* Compile and run the test with `--debug-memory`.
* Verify that the old compiler binary outputs a memory leak report.
* Verify that after the fix in `pkg/topology`, the test passes with exactly `0 leaked bytes` and all temporary allocations are cleanly dropped.
