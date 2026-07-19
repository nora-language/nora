# Walkthrough: Fiber Scheduler Cleanup, Concurrency, and Allocator Race Fixes

## Overview
This walkthrough summarizes the changes made to resolve intermittent segmentation faults, list corruption, and deadlock hangs during scheduler termination and concurrent fiber execution. We addressed a critical data race in the static/untracked allocator initialization, target-feature mismatch issues on AVX-only architectures, scheduler lost wakeups, active/terminated fiber list pointer corruptions, and main fiber thread-pinning issues.

## Changes

### 1. Eager Initialization of Untracked Allocator Mutex (`std/runtime/memory.c`)
- **Problem**: `g_untracked_lock` was initialized lazily within `nr_malloc_untracked()` and `nr_free_untracked()`. Under heavy concurrent fiber spawns, multiple worker threads called these functions simultaneously, causing a data race on `NR_MUTEX_INIT(&g_untracked_lock)`. This corrupted the mutex state, leading to deadlocks and race conditions.
- **Solution**: Moved `g_untracked_lock` and `g_untracked_lock_init` declarations to the top of `memory.c` and initialized the mutex eagerly inside `nr_mem_init()`, which runs on the main thread before starting any workers. Removed all lazy-initialization checks from allocation and deallocation paths.

### 2. SIMD Test Compatibility Check (`pkg/cmd/nora/compiler_test.go`)
- **Problem**: `simd_test.nr` requires AVX2/AVX-512 features (specifically instructions like gathers/scatters), but the test machine's CPU only supports AVX. Hardcoding AVX2 and FMA target features in `compiler_test.go` caused Clang to emit AVX2 instructions, resulting in `signal: illegal instruction (core dumped)` at runtime.
- **Solution**: Added a dynamic feature check at the start of the integration test runner to inspect `/proc/cpuinfo` for AVX2 support. If AVX2 is not supported by the host CPU, `simd_test.nr` is automatically skipped to prevent illegal instruction crashes.

### 3. Lost Wakeup Race Resolution (`std/runtime/fiber.c`)
- **Problem**: When a fiber yielded (parking) and another thread concurrently resumed it, a race occurred in `resume()`. If the CAS transition from `PARKED` to `READY` failed because the fiber was still parking, `resume_pending` was incremented. However, if the parking thread set the state to `PARKED` right before `resume()`'s second CAS check, `resume_pending` was decremented but the fiber was never queued, resulting in a lost wakeup and indefinite hang.
- **Solution**: Updated `resume()` such that if the state transitions from `PARKED` to `READY` via `atomic_compare_exchange_strong`, it is unconditionally pushed to the queue, and the `resume_pending` decrement operates independently.

### 4. Active/Terminated Fiber List Pointer Safety (`std/runtime/fiber.c`)
- **Problem**: When recycled fibers were popped from `g_terminated_fibers_head` and added back to `g_fibers_head` in `scheduler_spawn()`, `info->prev_global` was left pointing to a stale pointer from its previous lifetime. When the fiber terminated again, the stale `prev_global` pointer was dereferenced to update the list link, causing memory corruption and doubly-linked list corruption.
- **Solution**: Explicitly set `info->prev_global = NULL;` when prepending the recycled fiber to `g_fibers_head` inside `scheduler_spawn()`.

### 5. Main Fiber Thread Pinning (`std/runtime/fiber.c`)
- **Problem**: The main fiber (`"main"`) could be scheduled on any worker thread. When the main fiber yielded, its context (including worker stack pointers) was saved. When it was later resumed on a different thread (such as the main thread), it restored the other worker thread's stack context. This led to a scenario where multiple OS threads executed on the same stack space concurrently, corrupting stack parameters and causing `pthread_join()` to receive garbage stack addresses (such as `0xa` or `g_local_queues`), leading to SIGSEGV.
- **Solution**: Pinned the main fiber to Worker 0 (the Main Thread) in `scheduler_spawn` by checking if the fiber's name is `"main"` and pushing it to `g_pinned_queues[0]`.

## Verification & Testing
- Ran the full integration test suite via `go test -v ./pkg/cmd/nora`.
- Ran a stress-test loop of `build/release/spectralnorm 100` for 50 iterations; all executed successfully and stably.
- All integration tests completed successfully (`PASS` with exit code `0`).
- `simd_test.nr` is correctly skipped on non-AVX2 hosts, while all other concurrent, semantic, and compiler tests successfully pass.

