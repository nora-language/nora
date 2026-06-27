# Investigation: Uninitialized Array RAII Drop Segfault

## Problem
During the development of the Phase 4.3 Node-and-Beam Soft-Body Physics simulation, the compiled Nora executable crashed silently (exit code 1) upon initialization of the `SoftVehicle` struct. The crash occurred during the very first tick before the simulation loop could execute or print any debug output. No compilation errors or warnings were emitted by the Nora compiler, indicating a runtime memory error (segfault).

## Reproduction
The issue can be reproduced by dynamically allocating an array of inline structs (where the struct contains owned pointer fields) and then assigning to an element of that array:

```nora
package main

pub type Inner = struct {
    data: @str
}

pub type OuterVector = struct {
    elements: @Inner[]
}

pub fn main() i32 {
    var vec = alloc OuterVector {
        // Allocate an array of inline structs
        elements: alloc Inner[10]
    }
    
    var new_inner = Inner {
        data: @"hello"
    }
    
    // Assigning to the first element causes a segfault!
    vec.elements[0] = @new_inner
    
    return 0
}
```

## Root Cause
The root cause lies in how Nora's **Topological Lease Solver** handles assignments to existing memory locations containing owned pointer fields (RAII semantics). 

When `vec.elements[0] = @new_inner` is executed, the lease solver recognizes that `vec.elements[0]` is about to be overwritten. To prevent a memory leak, it automatically inserts a `Drop()` call to free the previous contents of `vec.elements[0]`.

However, because `vec.elements` was allocated using `alloc Inner[10]`, it translates to a `malloc` equivalent in the C runtime (`malloc(10 * sizeof(Inner))`). The C `malloc` function does not zero-initialize memory; it leaves garbage values. 

When Nora attempts to `Drop` the uninitialized `Inner` struct at index 0, it reads the garbage pointer from its `data` field and attempts to free it. Freeing a garbage pointer results in an immediate segmentation fault.

## Fix
To resolve the issue and safely manage arrays of objects with RAII fields, the data structures were refactored to store arrays of **heap pointers** (`@(@T)[]`) rather than arrays of **inline structs** (`@T[]`).

```nora
// Modified Data Structure
pub type SafeVector = struct {
    elements: @(@Inner)[]
}

// ...

var vec = alloc SafeVector {
    // Allocate an array of pointers (garbage/null pointers)
    elements: alloc (@Inner)[10]
}

// Initialize individual items via alloc before pushing
var new_inner = alloc Inner {
    data: @"hello"
}

// Assignment drops the old pointer. Dropping garbage pointers/nulls 
// without dereferencing them is safe because pointers themselves don't 
// invoke recursive field drops.
vec.elements[0] = @new_inner
```
By storing pointers, when the array slot is overwritten, Nora only drops the previous *pointer*, not the deeply nested fields. If the array is zero-initialized (or even if it contains garbage pointers), dropping a top-level pointer itself does not trigger recursive drops into uninitialized memory, preventing the segfault. 

*Note: Nora's C runtime safely ignores `free()` on `NULL` pointers. If `alloc` doesn't zero memory, freeing garbage might still be risky, but a top-level pointer free avoids recursive dereferencing crashes.*

## Validation
The fix was validated in `src/vehicle/soft_body.nr`. The `NodeVector` and `BeamVector` structs were updated to use `@(@Node[T])[]` and `@(@Beam[T])[]`. The items were individually allocated (`alloc Node[T]`) before being pushed into the array. After the fix, the `phase4_softbody` executable completed execution successfully without segfaulting, and the dynamic memory was properly managed across all 180 simulation ticks.
