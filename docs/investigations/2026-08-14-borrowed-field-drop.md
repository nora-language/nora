---
title: "Vector Reallocation Causes Use-After-Free via Array RAII Drops"
status: "Active"
date: "2026-08-14"
---

# Investigation Report: Vector Reallocation Causes Use-After-Free via Array RAII Drops

## Problem
When a generic `collections.Vector[T]` reallocates its internal array (e.g., inside `Push`), it copies elements from the old array to the new array using `unchecked_get` and `unchecked_set`. Then, it replaces the old array via `v.data = @new_data`.

Because `unchecked_get` bypasses the Topological Lease Solver's dependency tracking, the solver still believes the old array `v.data` owns all of its elements. Consequently, when `v.data` is overwritten, the solver automatically inserts a `PreDrop` for the old array. The generated C code then iterates over the old array and calls `nr_drop_T` on every element. 

Since the elements were already MOVED to `new_data`, this invokes their destructors (and frees inner memory) while the new array still holds pointers/references to them. This results in a double-free or use-after-free corruption for any `T` that owns memory (like structs or interfaces). 

This specifically broke `nora_gecs/scheduler.nr`, where adding a `TransformSystem` (via the `System` interface) to the `Scheduler`'s vector triggered a reallocation. The reallocation caused `TransformSystem`'s inner `reads` and `writes` `BitSet`s to be freed (`0xbaadf00d` magic), causing a `Panic: invalid array header` in `Scheduler.rebuild`.

## Reproduction
When code calls:
```nora
var sched = gecs.NewScheduler()
var sys1 = gecs.NewTransformSystem()
sched.AddSystem[gecs.TransformSystem](sys1)
// If AddSystem reallocates the internal vector, sys1 is moved to the new array, 
// but its original slot in the old array is dropped, destroying its inner state!
```
The compiler compiles the array assignment (`v.data = @new_data`) into C as:
```c
    gecs_System* _old = v->data;
    v->data = new_data;
    // ...
    int _len = array_count(_old);
    for (int _i = 0; _i < _len; _i++) {
        if (_old[_i].vtable && _old[_i].vtable[0] != NULL) {
            _old[_i].vtable[0](_old[_i].data); // BUG: Drops the element that was MOVED to new_data!
        }
        // ... nr_free(data)
    }
    nr_free(_old);
```

## Root Cause
The issue is situated in the interaction between `std/collections/vector.nr` and the compiler's Topological Lease Solver (`pkg/topology`).
The lease solver enforces RAII drops on arrays when they are overwritten or go out of scope. However, `Vector` manually manages memory and moves elements out of the array before it is dropped. Because `Vector` uses `unchecked_get`, the compiler does not realize the elements were moved out, so it drops them again.

## Fix Plan
There are two potential fixes:
1. **Compiler Intrinsic (`unchecked_free`)**: Add a new built-in function to the compiler (e.g., `unchecked_free(v.data)`) that frees the array allocation directly *without* iterating over and dropping its elements. Update `collections.Vector` to use this intrinsic instead of letting RAII drop the array.
2. **Topological Solver Awareness**: Make `unchecked_get` properly flag the source array slot as "moved" so the solver skips it during array drop. However, this is difficult since array drops in C iterate over the entire array length based on `array_count(_old)`.

**Recommended Fix:** Implement `unchecked_free` in the compiler (`pkg/hir` and `pkg/codegen`) and update `std/collections/vector.nr` to use it when reallocating or destroying vectors.

## Validation
1. Create a test case in `pkg/cmd/test/fail_vector_realloc_drop` that creates a struct with allocated memory, adds it to a `Vector`, forces a reallocation, and checks if the memory is freed.
2. Verify the test crashes with `baadf00d` or `invalid array header` on `master`.
3. Implement `unchecked_free` in the compiler and use it in `std/collections/vector.nr`.
4. Validate that the test passes and `nora_gecs` runs without memory panics.
