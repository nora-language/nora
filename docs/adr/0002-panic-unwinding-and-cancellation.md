# Architecture Decision Record: Concurrency Panic Unwinding & Cancellation

**Status:** Accepted
**Date:** 2026-06-06

## Context
Nora's M:N scheduler requires robust mechanisms for handling edge cases in concurrent programming:
1. **Panic Sandboxing**: Currently, if a fiber panics (e.g., array out-of-bounds), the `nr_panic` function calls `exit(1)`. This forcefully terminates the entire OS process, making it impossible to write a resilient Nora web server. We must isolate panics to the fiber level.
2. **Context Cancellation**: Nora lacks a native cancellation token system to gracefully terminate a tree of asynchronous tasks. Without this, canceled background tasks continue to consume CPU and memory, leaking resources.

## Decision

We will implement a Two-Pillar Strategy for resilient concurrency:

### 1. setjmp/longjmp Panic Isolation
We will modify the C11 runtime (`fiber.c`) and the compiler's code-generation (`pkg/codegen/expressions.go` and `hir_codegen.go`) to utilize standard C `setjmp` and `longjmp` for fiber panic unwinding.
- When a fiber starts, the compiler-generated `__spawn_wrapper` will push a `setjmp` recovery point to the fiber's internal state.
- If `nr_panic` is triggered, it will inspect the current fiber. Instead of terminating the OS thread, it will `longjmp` back to the fiber's wrapper.
- The wrapper will catch the panic, safely free its tracked arguments, and terminate the fiber.
- If the fiber was spawned inside a `scope`, the panic state will be propagated to the parent thread's `WaitGroup`, causing the parent's `Wait()` call to re-panic, correctly cascading the crash up the structured concurrency tree.

### 2. std/context Package
Instead of adding opaque compiler "magic", we will implement an explicit `context` library in the standard library.
- The `context` package will provide `Context` structs using atomic state and channels to broadcast cancellation signals to an arbitrary number of listening fibers.
- Workers will cooperatively check `ctx.Done()` or `ctx.Err()`.
- Future iterations of the compiler may allow syntax like `scope(ctx) { ... }` to implicitly wire cancellation to all spawns within the block, but initially, it will be explicit via parameter passing.

## Consequences
- **Positive:** A single buggy HTTP handler cannot crash a Nora web server anymore.
- **Positive:** Nora gains production-grade gracefulness for network timeouts and cancellations.
- **Negative:** `setjmp` comes with a tiny initialization overhead per fiber creation (saving CPU registers). However, since `setjmp` is only called once at fiber startup, the runtime performance impact is virtually zero.
