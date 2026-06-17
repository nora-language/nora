# Investigation: Closure Use-After-Free Vulnerability

Status: Completed

## Problem
A critical Use-After-Free (UAF) vulnerability was discovered in the Nora compiler's handling of closures and pointers. While closures capturing raw pointers (`#i32`, `@T`) were partially mitigated, closures capturing *structs containing pointers* were entirely unhandled by escape analysis rules. When a closure captured a struct with an embedded lease and then escaped its scope (e.g. via `return`, variable assignment, or `spawn`), it could execute and dereference the pointer long after the original stack frame had been destroyed, causing unpredictable memory corruption and potential crashes.

## Reproduction
We demonstrated the issue by wrapping a lease in a struct and returning a closure that captures the struct:
```nora
pub type System = struct {
    database: #i32
}

fn bad_return() fn() void {
    var data = 999
    var sys = System { database: #data }
    
    // Captures sys, which holds a lease pointing to data!
    var closure = fn() void {
        var local_val: #i32 = sys.database
        // local_val now points to destroyed stack memory!
    }
    
    // The closure escapes, taking the dead pointer with it
    return closure 
}
```

When called, `bad_return` allocates `data` on the stack. The closure captures `sys` by value, copying the pointer to `data` into its own struct environment. The function then returns the closure to the caller, tearing down `data`'s stack frame. When the caller invokes the returned closure, `sys.database` is dereferenced, causing a Use-After-Free.

## Root Cause
The root cause was twofold:
1. **Shallow Pointer Checks**: The compiler's escape analysis previously only looked for direct pointer types being captured. It did not recursively inspect fields of `StructType`, `SumType`, or elements of `ChanType` to determine if a lease was embedded within the data structure.
2. **Missing Escape Boundaries**: The compiler failed to reliably detect when a closure object escaped the scope that owned its captured variables. There were missing validations at `ReturnStatement`, `VarStatement` (assigning to global scopes), `AssignmentStatement`, and `SpawnExpression` boundaries.

## Fix
1. **Recursive Lease Detection (`hasLease`)**: Implemented a recursive type inspector in the semantic analyzer that dives into `StructType` fields, `SumType` variants, `MapType`, `ListType`, and `ChanType` to find any nested `Leased` pointers.
2. **Tainted Closure Flag (`CapturesLease`)**: When a lambda is analyzed, its captured variables are checked via `hasLease`. If any captured variable contains a lease, the resulting `FunctionType` is flagged with `CapturesLease = true` (making it a "restricted closure").
3. **Restricted Closure Escaping Blocks**: We implemented `hasRestrictedClosure(t types.NRType)` to detect restricted closures and added hard compiler errors when a restricted closure attempts to cross safe boundaries:
   - Returning from functions (`AnalyzeReturnStatement`)
   - Assigning to variables in the `Package` or `Global` scope (`AnalyzeAssignmentStatement`, `AnalyzeVarStatement`)
   - Spawning onto new fibers (`AnalyzeSpawnExpression`)

## Validation
Created the `escape_analysis` integration test suite under `pkg/cmd/test/escape_analysis`. The suite includes both positive tests for safe, local usage of capturing closures, and negative tests proving the compiler correctly emits semantic errors when attempting to escape a lease via struct wrappers, returns, globals, or spawns. All tests successfully passed via the `go test` harness.
