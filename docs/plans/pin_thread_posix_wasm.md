# PinThread and UnpinThread Implementation for POSIX and WASM

## Status
Planned

## Metadata
- **Author:** Antigravity
- **Date:** 2026-07-14
- **Components:** `std/runtime`

## Goal
To implement `PinThread` and `UnpinThread` in `std/runtime/fiber.c` for POSIX (Linux/macOS) and WASM (Native and Emscripten) platforms, bringing them to parity with the existing Windows implementation. This allows fibers to be pinned to specific OS threads during execution, which is crucial for thread-local state and certain C API integrations (e.g. OpenGL).

## Affected Compiler Components
- `std/runtime/fiber.c`: The core C runtime file implementing the cooperative fiber scheduler.

## Implementation Checklist
### POSIX (Linux/macOS) Implementation
- [ ] Add `int pinned_worker_id;` to `fiber_info_t` struct (around line 1040).
- [ ] Add `global_queue_t g_pinned_queues[MAX_WORKERS];` to the global state variables (around line 1205).
- [ ] In `scheduler_init`, initialize the pinned queues via `queue_init(&g_pinned_queues[i]);` in a loop.
- [ ] In `worker_loop` and `scheduler_run_loop`, check and pop from `g_pinned_queues[worker_id]` before popping from `g_local_queues`.
- [ ] In `resume`, check `info->pinned_worker_id`. If `>= 0`, push to `g_pinned_queues[info->pinned_worker_id]` instead of `g_local_queues` or `g_queue`.
- [ ] In `scheduler_spawn`, initialize `info->pinned_worker_id = -1;`.
- [ ] Implement `nr_fiber_pin_thread()` which gets the current fiber via `GetFiberData()` and sets `pinned_worker_id = worker_id`.
- [ ] Implement `nr_fiber_unpin_thread()` which gets the current fiber and sets `pinned_worker_id = -1`.

### Native WASM (`__wasm__`) Implementation
- [ ] Add `int pinned_worker_id;` to `fiber_info_t` struct (around line 771).
- [ ] In `scheduler_spawn`, initialize `info->pinned_worker_id = -1;`.
- [ ] Implement dummy `nr_fiber_pin_thread()` and `nr_fiber_unpin_thread()` (or assign `0` / `-1` since it's single threaded with `MAX_WORKERS 1`).

### Emscripten WASM (`__EMSCRIPTEN__`) Implementation
- [ ] Add `int pinned_worker_id;` to `fiber_info_t` struct (around line 1719).
- [ ] In `scheduler_spawn`, initialize `info->pinned_worker_id = -1;`.
- [ ] Implement dummy `nr_fiber_pin_thread()` and `nr_fiber_unpin_thread()`.

## Test Plan
- Run existing examples that use `PinThread`, like `glfw_test` or `engine_demo`, on Linux/macOS to ensure they compile and work successfully without crashing.
- Write an integration test in `pkg/cmd/test/` to spawn a fiber, pin it, yield, and verify it resumes on the exact same thread (e.g. by checking thread ID via a small C FFI snippet), then unpin it and verify it can migrate.
- Verify memory leak checks (`nr_mem_report()`) show no regressions.

## Risks
- POSIX scheduler performance might slightly drop due to an extra queue check, though it's negligible since pinned fibers are rare.
- Thread starvation: Pinned fibers rely on a specific thread waking up to process them. If that thread is busy or blocked, the pinned fiber will stall. Ensure pinned queues correctly signal `sem_post(&g_worker_sem)` just like regular queues.

## Completion Criteria
- Code for POSIX, Native WASM, and Emscripten successfully parses, compiles, and links.
- No compiler diagnostics or runtime crashes when `PinThread` is used on Linux.
- Integration tests pass successfully.
