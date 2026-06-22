# Specification: Parallel Iterators (`ParMap` / `ParIter`)

**Status:** Proposed

## 1. Motivation
In Nora, the Topological Lease Solver correctly prevents data races by banning mutable leases (`&T`) from crossing fiber `spawn` boundaries. While this guarantees absolute safety, it makes data-parallel operations (like parallel array modifications in physics engines, ECS frameworks, and rendering systems) cumbersome. Developers must move ownership (`@T`) of array elements through channels, causing heavy boilerplate.

We propose a native Parallel Iterator standard library feature (`std/collections/slice.ParMap`) that abstracts away the boilerplate, utilizing internal `unsafe` disjoint pointers to safely parallelize operations across arrays.

## 2. Syntax & Usage

```nora
import "std/collections/slice"

// A slice of bodies we want to mutate in parallel
var bodies = ps.world.bodies

// ParMap automatically chunks the array, spawns fibers inside a secure scope,
// and passes disjoint mutable leases to the lambda.
bodies.ParMap(fn(b: &Body[T]) {
    b.position = b.position.Add[T](#velocity)
})
```

## 3. Semantics & Memory Safety

1. **Structured Concurrency**: `ParMap` uses an internal `scope { ... }` block. It blocks the calling thread until all parallel chunks have finished executing. This guarantees the fibers never outlive the original array.
2. **Disjoint Mutability Guarantee**: The implementation of `ParMap` in the standard library will use `unsafe` pointer arithmetic to divide the slice into non-overlapping contiguous sub-slices (chunks). 
3. **No Aliasing**: Because the `ParMap` function guarantees that no two fibers receive a reference to the same index, the compiler can safely suspend its aliasing checks and allow the mutable leases (`&T`) to be handed to the closures.

## 4. Implementation Details (Standard Library)
The implementation does not require massive modifications to the Nora compiler frontend. It primarily relies on existing `unsafe` blocks and `spawn` capabilities within `std`.

```nora
// Pseudocode for std/collections/slice implementation
pub fn (s: &Slice[T]) ParMap(closure: fn(&T)) {
    var num_cores = runtime.NumCores()
    var chunk_size = s.len / num_cores
    
    var wg = sync.NewWaitGroup()
    wg.Add(num_cores)
    
    scope {
        var i = 0
        while (i < num_cores) {
            // Unsafe split to guarantee disjoint bounds
            var sub_slice = unsafe_split_slice(s, i * chunk_size, chunk_size)
            
            spawn map_worker[T](sub_slice, closure, #wg)
            i = i + 1
        }
    }
    
    wg.Wait()
}
```

## 5. Errors & Diagnostics
- If a user attempts to manually `spawn worker(&arr[i])`, the compiler will continue to emit: `Error: cannot pass mutable lease (write) across fiber boundary`.
- The compiler diagnostic help text should be updated:
  `help: To mutate an array in parallel, use the .ParMap() method instead of manually spawning fibers.`
