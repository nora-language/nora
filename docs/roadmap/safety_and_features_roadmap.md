# Nora Language: Safety and Features Roadmap

Following a comprehensive assessment of the compiler codebase (`pkg/semantic/analyzer.go`), the standard library (`std`), and advanced implementations like the `gecs` ECS port, we have updated the roadmap. Features like the Try Operator (`?`), Option/Result patterns, and Exhaustive Pattern Matching are already fully implemented. 

The focus is now shifted to resolving critical safety holes in concurrency and major ergonomic limitations in generic programming.

## 1. Concurrency Safety via Marker Traits (`Send`/`Sync`)
**Goal:** Eliminate unsafe hardcoded compiler exceptions and formally guarantee Thread-Safety for shared state.
**Current State:** To prevent Data Races, the compiler forbids passing read leases (`#T`) across fiber boundaries. However, to allow synchronization primitives to work, the `analyzer.go` uses a dangerous hardcoded string check: `strings.Contains(name, "WaitGroup") || strings.Contains(name, "Mutex")`. This allows any user-defined struct with "Mutex" in its name to bypass fiber safety checks.
**Proposal:** 
- Introduce compiler-recognized marker interfaces (e.g., `ThreadSafe`, or `Send`/`Sync` equivalents).
- Safely derive these markers for structs only if all their fields are also `ThreadSafe`.
- Completely remove the string-matching hacks from `pkg/semantic/analyzer.go`.

## 2. Primitive Type Extensions (True Typeclasses)
**Goal:** Enable zero-cost abstractions and ergonomic generic collections.
**Current State:** `collections.HashMap` requires the user to pass manual function pointers `hash_fn` and `eq_fn` upon initialization. This is because **Nora does not allow defining methods on primitive types** (e.g., `i32`, `str`). Consequently, primitives cannot satisfy generic constraints like `[T: Hashable]`, breaking the usability of Typeclasses.
**Proposal:**
- Implement the ability to attach methods to primitive and foreign types (e.g., `fn (self: #i32) Hash() i32`).
- Refactor `std/collections/map.nr` and other generic data structures to use bounded generic constraints (`[K: Hashable + Equatable]`) rather than requiring manual function pointers.

## 3. Structured Concurrency and Fiber Lifetime Bounds
**Goal:** Prevent Use-After-Free in asynchronous task execution.
**Current State:** While the compiler stops arbitrary read-leases from crossing fiber boundaries, "exempt" types (like `#WaitGroup`) can be freely passed. Nora lacks lifetime bounding (like Rust's `'static` bound), meaning a parent fiber can spawn a child fiber with `#wg` and then return, dropping `wg` while the child fiber still references it.
**Proposal:**
- Introduce **Structured Concurrency** (e.g., `Nursery` or `Scope` blocks).
- Enforce that fibers capturing local stack leases (even thread-safe ones) must be spawned within a scope that guarantees the fibers are joined *before* the variables go out of scope.

## 4. Formalization of Operator Overloading
**Goal:** Provide safe, predictable, and consistent syntax for custom types.
**Current State:** Nora currently allows custom equality checks via generic `a == b` which dynamically resolves if a method `fn eq(self: #T, other: #T) bool` is defined on the struct. This relies on "magic" method naming rather than formal interfaces.
**Proposal:**
- Tie operator overloading directly to standard library interfaces (e.g., `std.cmp.Equatable`, `std.cmp.Comparable`, `std.ops.Addable`).
- Ensure the compiler verifies interface satisfaction before emitting the overloaded C-operations.

## 5. Strict Escape Analysis & Closure Freezing
**Goal:** Ensure memory safety around closures and captured variables.
**Current State:** Nora correctly enforces "frozen violations" (preventing reassignment of variables captured by active closures) and Topological Lease Solving for RAII.
**Proposal:**
- Enhance the Escape Analysis to statically ensure closures capturing leases do not escape the lifecycle of the host variables.

## 6. Bounds Check Analysis & Elimination
**Goal:** Maximize raw execution speed for tight loops without sacrificing array safety.
**Current State:** The compiler relies heavily on runtime bounds checking (`array_bounds_check` in C codegen) to prevent Buffer Overflows.
**Proposal:**
- Implement a static Range Analysis pass in the semantic analyzer.
- Prove at compile-time when an index is mathematically guaranteed to be within bounds (e.g., inside a bounded `for i in 0..vec.len()` loop).
- Safely omit the runtime bounds check in the emitted C code for proven cases, leading to zero-cost abstractions.
