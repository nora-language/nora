# Compiler Investigation: Topological Solver Fails to Drop Anonymous Temporaries

## Status
Completed

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
The root cause resided in two places across the compiler pipeline:
1. **Topological Lease Solver (`pkg/topology`)**: Previously, the solver only tracked temporary r-values if they were inline lambda expressions (`ast.LambdaExpression`). Other r-values returning owned types (`ast.CallExpression`, `ast.StructInstantiation`, `ast.ArrayLiteral`) that were unconsumed by an outer move operation were ignored during block analysis.
2. **HIR Lowering (`pkg/hir`)**: Even if drop instructions were generated for an arbitrary expression, the HIR lowering phase did not assign a temporary stack variable (`Alloca` + `Store`) for non-lambda expressions. As a result, when emitting HIR drop instructions, there was no variable symbol to target.

## Fix
1. **`pkg/topology/solver.go`**: Generalized r-value drop scheduling. Replaced `findUnconsumedLambdas` with `findUnconsumedRValues`, checking if any expression node (`CallExpression`, `StructInstantiation`, `ArrayLiteral`) has an owned return type (`isOwnedRValueType`) and is unconsumed by its parent node. Scheduled `DropInfo{Expr: expr}` at the end of the statement for all unconsumed r-values.
2. **`pkg/hir/instruction.go`**: Added `Expr ast.Expression` field to `hir.Drop` instruction and updated `String()` representation.
3. **`pkg/hir/lower.go`**: Added `ExprTemps` and `UnconsumedTemps` to `Lowerer`. When `NewLowerer` initializes, it populates `UnconsumedTemps` from the solver's scheduled `Drops` and `PreDrops`. Wrapped `lowerExpression` so that when lowering an unconsumed temporary expression, it allocates a stack temporary (`_hir_tmp_N`), stores the value into it, records the symbol in `ExprTemps`, and emits `Drop(sym)` when the statement finishes.

## Validation
* Created positive reproduction integration test: `pkg/cmd/test/repro_anonymous_temporary_leak/main.nr`.
* Executed `nora run --debug-memory pkg/cmd/test/repro_anonymous_temporary_leak/main.nr` before the fix: confirmed 12-byte memory leak reported.
* Executed `nora run --debug-memory pkg/cmd/test/repro_anonymous_temporary_leak/main.nr` after the fix with cleared cache: confirmed 0 leaked bytes and clean exit.
