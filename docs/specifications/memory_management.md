# Memory Management & Topological Lease Solver

## Overview

Nora achieves compile-time memory safety and resource management without a garbage collector (GC). The language relies on a static **Topological Lease Solver** paired with **Contextual Lifecycle Leases**. Resources are tied to scopes and explicitly transferred, borrowed, or automatically dropped via RAII principles.

## Motivation

Traditional systems languages require either manual memory management (`malloc`/`free` in C) leading to leaks and use-after-free bugs, or strict borrow checkers (Rust) that can impose a steep learning curve. Garbage collectors (Go, Java) introduce runtime overhead and unpredictable pauses. Nora's Topological Lease Solver eliminates runtime GC overhead while automatically inserting resource cleanup routines based on lease dependency graphs.

## Syntax & Lease Operators

Nora uses explicit markers to define ownership and borrowing states at function boundaries and struct definitions.

### 1. The `alloc` Keyword
All memory allocations (both heap and complex stack objects requiring deterministic drop tracking) must go through the explicit `alloc` keyword.

```nora
var user = alloc User { name: "Alice" }
var buffer: alloc i32[1024]
```

### 2. Owned Move (`@`)
Transfers ownership (consumption) of a value or variable. Once moved, the original variable can no longer be used.

```nora
pub fn TakeOwnership(data: @String) {
    // data is consumed here
}
```
In structs, `@` defines a field that physically owns its contents.

### 3. Read-Only Borrow (`#`)
Represents a read-only lease (immutable borrow). The function or scope may read the data but cannot modify it or transfer its ownership.

```nora
pub fn ReadData(data: #String) {
    // Can read, cannot modify
}
```

### 4. Mutable Borrow (`&`)
Represents a mutable lease (read-write exclusive borrow). 

```nora
pub fn ModifyData(data: &String) {
    // Can mutate the original String
}
```

### 5. Manual Pinning (`pin`)
For fine-grained manual overrides of the topological solver, developers can use the `pin` keyword. This forces a resource's lease to remain alive for the remainder of the current scope, circumventing any implicit moves or drops that might otherwise occur. This is exceptionally useful when passing resources to asynchronous C functions via FFI that retain a pointer across an asynchronous boundary.

```nora
var r = alloc Resource { id: 42 }
pin r
do_async_c_stuff(r) // FFI call that keeps the pointer internally
// r is guaranteed to not be dropped until the block ends
```

## Semantics & Topological Lease Solver

The Topological Lease Solver (`pkg/topology`) tracks variable lifetimes using a dependency graph. 

1.  **Birth & Lifecycle:** Tracks when a variable is instantiated (`alloc` or standard declaration).
2.  **Tracking:** The solver tracks assignments, moves, and pins (anchoring to block ends).
3.  **RAII Drop Insertion:** At the end of a block scope, the solver evaluates the dependency graph and automatically inserts `PreDrops`, `Drops`, and `TryDrops` for any un-moved, owned resources.

## Type & Lease Rules

1.  **Partial/Field Moves:** Moving a field out of a struct (e.g., `var id = @user.id`) prevents further usage of the now partially-moved `user` object.
2.  **Re-assignment:** If an owned variable is re-assigned, the solver triggers a `PreDrop` on the old value before the overwrite occurs to prevent leaks.
3.  **Moves:** Passing an owned variable to a function expecting an `@` argument consumes it.

## Examples

### Struct Ownership

```nora
pub type Container = struct {
    elements: @Vector[i32]
}

fn main() {
    var c = Container {
        elements: alloc Vector[i32]()
    }
    // `c.elements` is automatically dropped when `c` goes out of scope.
}
```

## Errors & Diagnostics

*   **Use After Move:** Attempting to use a variable after it has been moved via `@` yields a compilation error.
*   **Leak Detection:** During debug compilation (`--debug-memory`), any un-tracked drops are reported by the `nr_mem_report()` utility.
