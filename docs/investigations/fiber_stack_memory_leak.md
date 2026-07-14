# Investigation: Fiber Stack Memory Leak in Nora Runtime

## Status
**RESOLVED** — Fix applied in `std/runtime/fiber.c`

## Problem

When running programs that continuously spawn fibers in a loop (such as the `nora_wgpu` multi-threaded rendering example which spawns 4 fibers per frame), the application memory increased unboundedly. At 60 FPS, the memory footprint would grow by approximately 240MB per second.

When running with the `--debug-memory` flag, the Nora compiler reported **0 memory leaks**, which made the issue difficult to diagnose initially.

## Reproduction

```nora
pub fn main() {
    while true {
        spawn worker_fiber()
    }
}

fn worker_fiber() {
    // Do some work and exit
}
```
Running this code would quickly exhaust system memory.

## Root Cause

The memory leak stemmed from two interacting systems:

1. **Fiber Stack Allocation Bypass:** Windows OS fibers are created using `CreateFiber()`, which allocates a 1MB stack directly from the OS. Because this allocation occurs inside the native OS layer and not through Nora's standard memory allocator (`nr_malloc`), the `--debug-memory` tracker was completely blind to it.
2. **Delayed Cleanup of Terminated Fibers:** When a Nora fiber completes execution, it calls `park()` and its control block is appended to a global linked list: `g_terminated_fibers_head`. However, the only place where this list was ever processed and freed (via `DeleteFiber()` and `free()`) was inside `scheduler_cleanup()`, which is only called when the *entire application terminates*. 

Because `g_terminated_fibers_head` was never cleaned up while the application was running, every spawned fiber's 1MB stack was kept alive in memory forever, even after the fiber had successfully terminated.

## Fix

We applied a patch to `std/runtime/fiber.c` to proactively clean up terminated fibers whenever a new fiber is spawned. By hooking into `scheduler_spawn()`, we guarantee that the memory footprint of dead fibers is regularly recycled during the exact operations that increase fiber counts.

```c
// std/runtime/fiber.c: scheduler_spawn()

void* scheduler_spawn(void (*fn)(void*), void* arg, const char* name, const char* file, int line) {
    // --- BEGIN PATCH ---
    NR_MUTEX_LOCK(&g_fiber_list_lock);
    fiber_info_t* curr = g_terminated_fibers_head;
    while (curr) {
        fiber_info_t* next = curr->next_global;
#ifdef _WIN32
        if (curr->handle) DeleteFiber(curr->handle); // Frees the 1MB OS stack
#endif
        free(curr); // Frees the Nora fiber control block
        curr = next;
    }
    g_terminated_fibers_head = NULL;
    NR_MUTEX_UNLOCK(&g_fiber_list_lock);
    // --- END PATCH ---

    fiber_info_t* info = (fiber_info_t*)malloc(sizeof(fiber_info_t) + sizeof(spawn_data_t));
    // ... continues with fiber initialization
}
```

## Validation

After applying the patch, running the `nora_wgpu` multi-threaded example (spawning 240 fibers per second) resulted in a completely flat and stable memory footprint. The memory leak is fully resolved.

## Affected Files
- `std/runtime/fiber.c`
