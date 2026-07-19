# Compiler Investigation Report: Fiber Scheduler Cleanup Crash

Status: Completed
Created: 2026-07-19
Updated: 2026-07-19
Author: Antigravity

## Purpose

Analyze the segmentation faults (`SIGSEGV` / `SIGABRT`) and hangs occurring intermittently during process exit or thread teardown under high-concurrency fiber execution (e.g. running `spectralnorm 100` stress test).

---

## Problem

Under release builds, stress-running concurrent programs with multiple worker threads resulted in random segmentation faults (exit code 139) or deadlock hangs inside `__pthread_clockjoin_ex` during compiler runtime shutdown (`scheduler_cleanup`).

The crash stack trace showed:
- Frame #0: `__GI___libc_free`
- Frame #4: `__GI___nptl_deallocate_stack`
- Frame #6: `__pthread_clockjoin_ex`
- Frame #8: `scheduler_cleanup` at `std/runtime/fiber.c`

Further, in some coredumps:
- `threadid` passed to `pthread_join` was corrupted (e.g., `0xFFFFFFFF` or pointing to BSS structures like `g_local_queues`).

---

## Reproduction

Build and run `examples/spectralnorm/simd/spectralnorm.nr` in a loop:
```bash
./nora build -r examples/spectralnorm/simd/spectralnorm.nr
for i in {1..100}; do ./build/release/spectralnorm 100 >/dev/null; done
```
This intermittently triggers a crash (usually within 2-5 runs) on multicore Linux hosts.

---

## Root Cause

Two critical concurrency issues in `std/runtime/fiber.c` were identified:

1. **XSAVE Area Buffer Overflow (Primary Cause of BSS/Thread ID Corruption)**:
   - Modern x86_64 CPUs running glibc save floating-point/vector registers (YMM/ZMM) in an extended context area (`XSAVE` area) that can exceed 2.6 KB when AVX/AVX-512 is active.
   - The thread and fiber context storage structures (`fiber_info_t` and `padded_ucontext_t`) allocated only `512 bytes` of padding after the `ucontext_t context` field.
   - Calling `swapcontext` wrote up to 2,688 bytes into the 512-byte buffer, overflowing by 2+ KB. This directly corrupted adjacent structures in the BSS section, such as `g_worker_threads[MAX_WORKERS]`, overwriting valid `pthread_t` handles with pointers like `&g_local_queues[0]` or `-1` (`0xFFFFFFFF`). When `scheduler_cleanup` joined these handles, it crashed inside glibc's pthread library.

2. **Unprotected `park()` Dereferencing `NULL`**:
   - `park()` in `fiber.c` did not verify if `GetFiberData()` returned `NULL`.
   - If `GetFiberData()` returned `NULL` (e.g., due to thread-local `worker_id` mismatch or uninitialized worker structures during early initialization/late teardown), `park()` would proceed to execute `swapcontext(&info->context, ...)` with a null-offset pointer, leading to a segfault.

3. **Lack of Mutex/Flag Protection Against Re-entrant/Concurrent Cleanup**:
   - `scheduler_cleanup()` is called at the end of the `main()` function execution. Under some races or double calls, the same thread IDs could be joined multiple times, or lock destruction could occur concurrently, leading to undefined behaviors and hangs.

---

## Fix

1. **Increase XSAVE Padding**:
   - Increased the padding in `padded_ucontext_t` and `fiber_info_t` to `4096 bytes` (`_Alignas(64) char _context_padding[4096]`), which safely accommodates AVX-512 and AMX registers.
   
2. **Add NULL Guard in `park()`**:
   - Added a safe guard at the top of `park()` to return immediately if `info` is `NULL`.

3. **Atomic Re-entrancy Protection for `scheduler_cleanup()`**:
   - Introduced an atomic `g_cleaned_up` flag at the top of `scheduler_cleanup()` to ensure the teardown routine runs exactly once and avoids concurrent/re-entrant cleanup attempts.
