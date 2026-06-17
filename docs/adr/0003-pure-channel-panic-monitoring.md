# Architecture Decision Record: Pure Channel Panic Monitoring

**Status:** Accepted
**Date:** 2026-06-06

## Context
Nora emphasizes strict ownership, data-race freedom, and structured concurrency. When a detached background fiber panics (e.g., due to an unexpected connection reset or out-of-bounds error), we need a safe mechanism to catch that panic without terminating the OS process. 
Initially, we considered implementing a generic `JoinHandle` to act as an awaitable promise. However, standard `JoinHandle` patterns encourage polling and manual `await` checks, which don't scale well to dynamic aggregators like the `std/nursery`. A `nursery` needs to multiplex thousands of dynamic fiber completions, but Nora's `select` statement requires static cases.

## Decision
We rejected the `JoinHandle` implementation in favor of a "Pure Channel" monitoring syntax directly integrated into the language. 

The `spawn` expression is extended to accept an optional monitor channel:
```nora
var monitor = make(chan[str], 100)
spawn(monitor) risky_background_task()
```

When `spawn(monitor)` is invoked:
1. The compiler lowers it into a `Spawn` HIR instruction storing the channel operand.
2. The C11 code generator injects the `monitor_chan` directly into the `__spawn_args_N` internal struct.
3. The generated `__spawn_wrapper` wraps the fiber execution in a `setjmp` block.
4. If a panic occurs, the runtime uses `longjmp` back to the wrapper, which reads the string panic reason and automatically issues a zero-copy `channel_send` to the monitor channel before terminating safely.

## Alternatives Considered
- **`JoinHandle`:** Rejected because dynamic multiplexing (selecting over an array of `JoinHandle`s) is highly inefficient and breaks Nora's static `select` paradigm.
- **Optional Function Parameters (`spawn fn(monitor)`):** Rejected because Nora specifically omits optional parameters from the language to keep the type system strict. Extending `spawn` as a core language construct cleanly sidesteps function signature bloat.

## Consequences
- **Positive:** Enforces the Fan-In concurrency pattern. The `std/nursery` can aggregate thousands of fiber completions and panics into a single, highly efficient `completion_chan` with zero dynamic allocations.
- **Positive:** `spawn` remains an expression, fitting cleanly into the language grammar.
- **Neutral:** `monitor_chan` is strictly required to be `chan[str]`. Future iterations may extend this to typed error enums.
