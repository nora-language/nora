# Standard Library: Runtime

## Overview

The `std/runtime` package exposes internal details of the Nora language runtime, specifically interacting with the Cooperative Fiber Scheduler, Memory Allocator, and Garbage Collection metrics (if applicable during fallback modes). Most applications will never need to interact with this package directly, as fibers and memory are handled transparently by keywords (`spawn`, `alloc`).

## Core Functions

### Fiber Control
*   `Yield()`: Voluntarily yields execution of the current fiber back to the scheduler, allowing other fibers queued on the OS thread to execute. This is normally injected automatically (`NR_COOPERATIVE_YIELD_CHECKPOINT`), but can be manually triggered in tight compute-heavy loops.
*   `NumCPU() i32`: Returns the number of logical CPUs usable by the current process.
*   `NumFibers() i32`: Returns the number of currently active fibers.
*   `LockOSThread()`: Wires the calling fiber to its current operating system thread. The calling fiber will always execute in that thread, and no other fiber will execute in it, until the calling fiber has made as many calls to `UnlockOSThread` as to `LockOSThread`. Useful for FFI/C-interop relying on Thread-Local Storage (TLS).

### Metrics
*   `MemStats`: A structure containing detailed metrics about memory allocated by the Nora runtime.
*   `ReadMemStats(m: &MemStats)`: Populates `m` with the latest memory allocation statistics.

## Internal Mechanisms

The runtime package directly interfaces with the underlying C11 core (`std/runtime/fiber.c` and `std/runtime/scheduler.c`). Because Nora is designed for zero-cost abstraction, calling `Yield()` is compiled down directly to a fast context switch (using `CreateFiber`/`SwitchToFiber` on Windows, or `ucontext` on POSIX).
