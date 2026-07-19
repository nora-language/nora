# Fix Fiber Scheduler Cleanup and Park/Yield Plan

## Status
Completed

## Metadata
- **Author:** Antigravity
- **Date:** 2026-07-19
- **Components:** `std/runtime/fiber.c`
- **Related Investigations:** `docs/investigations/fiber_scheduler_cleanup_crash.md`

## Goal
To resolve intermittent segmentation faults and deadlock hangs occurring during concurrent fiber scheduler termination. We will make the scheduler cleanup thread-safe and re-entrant safe, add NULL guards in context-switching routines (`park()`), and verify stability under high stress-testing concurrency.

## Affected Compiler Components
- `std/runtime/fiber.c`:
  - `park()` function (add NULL checks).
  - `scheduler_cleanup()` (make it re-entrant / double-call safe).

## Implementation Checklist

### 1. Add NULL Guard in `park()`
- [x] In `fiber.c` inside POSIX `park()`, verify `info` returned from `GetFiberData()`. If `NULL`, return immediately to prevent calling `swapcontext` with a null offset.

### 2. Implement Re-entrancy Protection in `scheduler_cleanup()`
- [x] Declare a static/atomic flag `g_cleaned_up` inside `fiber.c`.
- [x] Use atomic CAS/exchange on `g_cleaned_up` at the beginning of `scheduler_cleanup()` to ensure the teardown runs exactly once.

### 3. Add Concurrency/Cleanup Regression Test
- [x] Create a dedicated test folder under `pkg/cmd/test/scheduler_concurrency_cleanup_test/`.
- [x] Implement a test that spawns a large number of fibers, performs basic concurrent synchronization, and finishes, ensuring cleanup transitions are executed correctly.
- [x] Run stress tests in loops to verify that zero memory errors or hangs are present.

## Test Plan
1. **Stress Test Integration:** Run the N=100 standalone loop test on `spectralnorm` to ensure the previous SIGSEGV does not reproduce.
2. **Regression Test:** Create and run `pkg/cmd/test/scheduler_concurrency_cleanup_test/`.
