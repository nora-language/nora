# Compiler Investigation: Topological Solver Leaks on Implicit Deref to Value Fields

## Status
Completed

## Problem
When a generic function returning an owned pointer (e.g., `collections.NewVector[T]()` returning `@Vector[T]`) is assigned directly to a struct field of a value type (e.g., `Vector[T]`), the Nora compiler generates an implicit dereference (value copy). However, the Topological Lease Solver fails to insert a `Drop` for the original anonymous heap-allocated pointer wrapper (the `@Vector` itself), resulting in a leak of exactly 16 bytes per call.

## Reproduction
To reproduce the issue, compile and run code that assigns an allocated `@Vector` to a value-type field:

```nora
import "collections"

pub type IslandBatch[T] = struct {
    pairs: collections.Vector[T] // value type
}

pub fn TriggerLeak() {
    // BUG: The solver implicitly dereferences `@Vector` to copy into `pairs`.
    // The inner array data is moved, but the 16-byte `@Vector` pointer wrapper is never freed!
    var _new_island = IslandBatch[f64] {
        pairs: collections.NewVector[f64](4)
    }
}
```

Running this code with memory debugging (`nora run --debug-memory`) will output a leak report showing `Leak: 16 bytes` corresponding exactly to the `alloc Vector[T]` inside `NewVector`. Extracting the call into a named variable (`var _tmp = collections.NewVector()`) does not fix the issue because the solver interprets the assignment to the struct field as a "Move", transferring ownership and canceling the local variable's drop, without realizing only the inner value was copied.

## Root Cause
The root cause lies in the **Topological Lease Solver** (`pkg/topology`). 
When an `@T` (pointer) is assigned to a `T` (value type) struct field, the semantic lowering performs a shallow copy of the struct's fields. However, the lease solver treats this field assignment as a full ownership transfer (Move). Because it thinks the struct field now owns the pointer, it cancels the `Drop` for the local variable/anonymous temporary. But since the field is a value type, it cannot drop the pointer wrapper. Thus, the 16-byte wrapper allocation dangles forever.

## Fix / Workaround
1. **Physics Engine & Library Workaround (Applied):** All structures containing vectors or allocated structs (e.g. `PhysicsSystem.islands` and `IslandBatch.pairs`) have been updated to use `@collections.Vector` (pointer fields) instead of value types to prevent triggering implicit dereferencing moves.
2. **Compiler Enhancement (Implemented):** Updated `pkg/semantic/analyzer.go` to explicitly check struct literal field initialization types via `checkInterfaceCompatibility`, ensuring assignments of `@T` (move leases) to `T` (value struct fields) are rejected during semantic analysis.
3. **Topological Solver Fix (Implemented):** Corrected AST node recognition in `walkUnconsumedRValues` (`pkg/topology/solver.go`) from `StructInstantiation` to `StructLiteral` and added support for `FieldDefinition` nodes.

## Validation
- Created regression test `pkg/cmd/test/fail_repro_implicit_deref_leak_struct_literal/fail_repro_implicit_deref_leak_struct_literal.nr` verifying that assigning `@Vector` to a value struct field is rejected by semantic analysis.
- Verified with regression test `pkg/cmd/test/repro_implicit_deref_leak/repro.nr` running under `--debug-memory` that passing values by reference/pointer avoids leaks (`0 bytes leaked`).
- Verified tests pass with `go test -v ./pkg/cmd/nora -run TestCompilerWithTestFolder/repro_implicit_deref_leak` and `go test -v ./pkg/cmd/nora -run TestCompilerWithTestFolder/fail_repro_implicit_deref_leak_struct_literal`.
