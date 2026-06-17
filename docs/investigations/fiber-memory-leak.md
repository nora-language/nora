# Fiber Memory Leak Investigation

Status: Completed
Created: 2026-06-01
Updated: 2026-06-01
Author: Gemini

## Purpose

Analyze the "known infrastructure leak" reported during integration tests to determine if it is an actual resource/memory leak or acceptable infrastructure overhead.

---

## Problem

During integration test execution, Nora's memory leak detector identifies a leak of exactly 16 bytes for processes that spawn fibers. The test framework filters this out with a message:
`Ignoring known infrastructure leak in: <test_name>`

This investigation aims to determine:
1. Whether this leak represents an actual memory leak.
2. The root cause of the leak.
3. The implications on short-lived vs. long-running programs.
4. Proposed strategies to resolve or manage the leak.

---

## Reproduction

Any test case that invokes `spawn` or runs under the concurrent scheduler (such as `simple_concurrency.nr` or `http_test.nr`) will trigger the leak report if memory reporting is enabled (`nr_mem_report()`):

```
=== RUN   TestCompilerWithTestFolder/pkg\cmd\test\simple_concurrency\simple_concurrency.nr
...
Ignoring known infrastructure leak in: pkg\cmd\test\simple_concurrency\simple_concurrency.nr
```

---

## Findings

1. **`fiber_info_t` Allocation**:
   In `std/runtime/fiber.c`, the scheduler spawns fibers using `scheduler_spawn(...)`:
   ```c
   void* scheduler_spawn(void (*fn)(void*), void* arg, const char* name, const char* file, int line) {
       fiber_info_t* info = (fiber_info_t*)malloc(sizeof(fiber_info_t) + sizeof(spawn_data_t));
       ...
   }
   ```
   This allocates the `fiber_info_t` struct on the heap, along with `spawn_data_t` metadata (totalling 16+ bytes depending on the platform).

2. **Win32 Fiber Handles**:
   On Windows, `CreateFiber` is called to allocate the fiber stack:
   ```c
   info->handle = CreateFiber(0, (LPFIBER_START_ROUTINE)fiber_wrapper, info);
   ```

3. **Termination/Cleanup Absence**:
   When a fiber completes execution, its wrapper transitions its state to `DEAD`/`TERMINATED` (state `4`) and calls `park()` to switch back to the scheduler:
   ```c
   NR_ATOMIC_STORE(&info->state, 4); // TERMINATED
   park();
   ```
   Crucially, **neither `free(info)` nor `DeleteFiber(info->handle)` is ever called**!
   A running fiber cannot free its own struct or stack because doing so would destroy the active stack frame and context it is currently executing on, leading to an immediate crash.

4. **Scheduler Cleanup Omission**:
   Nora's `scheduler_cleanup()` function only stops workers and destroys global mutexes. It does not iterate over dead/terminated fibers to reclaim their memory or handles:
   ```c
   void scheduler_cleanup() {
       g_running = false;
       ReleaseSemaphore(g_worker_sem, num_workers, NULL);
       Sleep(50);
       CloseHandle(g_worker_sem);
       NR_MUTEX_DESTROY(&g_queue.lock);
   }
   ```

---

## Root Cause

The leak is an **actual resource leak** of both heap memory (`fiber_info_t`) and OS fiber stacks (allocated by `CreateFiber` / `ucontext` / `wasm_cont`). 

Because fibers cannot deallocate themselves on their own stack, they rely on a garbage collector or a dedicated scheduler sweep pass to reclaim terminated fiber stacks. Currently, Nora lacks a sweeping/reclaim mechanism in its run-loop and cleanup routines, resulting in persistent leaks for every fiber created.

---

## Fix and Recommendations

### Resolution

We have implemented **Strategy 2 (Explicit Sweep)**. The `scheduler_cleanup()` function now iterates through all global fibers and systematically frees the fiber structs and OS handles when the runtime shuts down. The integration test framework now strictly enforces a zero-leak policy.

### Long-term Production Fixes
To prevent memory exhaustion in long-running services (like the HTTP server), one of the following strategies should be implemented in `std/runtime/fiber.c`:

1. **Lazy Reclaim in the Scheduler Run-Loop**:
   Maintain a global queue of `terminated_fibers`. When a fiber completes, it pushes itself to the terminated queue before parking. The main scheduler loop can safely pop and free (`free(info)` and `DeleteFiber`) these dead fibers from the scheduler thread's stack.
   
2. **Explicit Sweep in `scheduler_cleanup()`**:
   In `scheduler_cleanup()`, iterate through the global fiber list and systematically free all fiber structs and handles.
