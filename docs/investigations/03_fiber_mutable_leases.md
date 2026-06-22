# Investigation: Fiber Disjoint Mutability and Mutable Leases

**Date:** 2026-06-22
**Status:** Completed
**Author:** Antigravity & User

## 1. Problem Statement
During the development of the `nora_physics` engine, we attempted to parallelize the dynamics integration phase. The goal was to spawn a fiber for every `Body[T]` in the `World` array to compute physics math (`v = v + g*dt`) in parallel.

```nora
pub fn (ps: &PhysicsSystem[T]) Step(dt: T) {
    scope {
        var i = 0
        while (i < num_bodies) {
            // Attempting to spawn a fiber and pass a mutable lease
            spawn worker_integrate[T](&ps.world.bodies[i].ptr)
            i = i + 1
        }
    }
}
```

The compiler rejected this with:
`Error: cannot pass mutable lease (write) across fiber boundary`

## 2. Root Cause Analysis
The Nora Topological Lease Solver guarantees memory safety by enforcing exclusive write access (`&T`) and preventing use-after-free conditions. 

When passing a reference across a `spawn` boundary, two issues arise:
1. **Lifetime Verification**: The compiler cannot guarantee that the spawned fiber will finish before the `World` array is destroyed.
2. **Disjoint Mutability (Aliasing)**: In a dynamic `while` loop, the compiler cannot statically prove that `&bodies[i]` and `&bodies[i+1]` point to disjoint memory locations. If they pointed to the same memory, multiple fibers would hold exclusive mutable leases (`&T`), causing a data race.

## 3. Review of the `[shared]` Attribute
We investigated the compiler's `cmd/test/` directory, specifically `shared_fiber_test.nr` and `fail_scope_normal_var.nr`. 
We confirmed that Nora supports structured concurrency: Fibers spawned inside a `scope { ... }` block are joined before the scope exits, solving the Lifetime Verification issue.

Furthermore, structs marked with `[shared]` (like `sync.WaitGroup` and `sync.Mutex`) are permitted to cross `spawn` boundaries. However, `[shared]` structs can only be passed via **Read Leases (`#T`)**. 

Because `Integrator` requires structural mutation of `Body.position`, it demands a **Mutable Lease (`&T`)**. Making `Body` `[shared]` does not bypass the exclusivity rule of mutable leases. Using `#Body` would require wrapping every body in an atomic or lock (e.g. `#Mutex[Body]`), which introduces unacceptable runtime overhead for a high-performance physics engine.

## 4. Conclusion & Workaround
Nora's strict enforcement of exclusive mutable leases is working exactly as intended to prevent data races. 
Because `spawn` only accepts owned moves (`@T`) or read leases (`#T`) of `[shared]` structs, achieving parallel physics integration requires removing bodies from the array (taking ownership), sending them to fibers, and funneling them back via zero-copy channels.

**To avoid massive boilerplate, we reverted the physics solver to a highly performant sequential loop.**

To natively solve this language-wide constraint for High-Performance Computing without forcing channel boilerplate, we propose adding a `Parallel Iterator` abstraction to the standard library. (See `docs/specifications/parallel_iterators.md`).
