# Asynchronous Execution: The `spawn` Keyword

## Title & Overview
The `scope` keyword provides structured concurrency, while `spawn` provides raw, detached asynchronous execution. The `spawn` keyword is Nora’s fundamental primitive for creating lightweight fibers. It allows a function or method to execute concurrently in the background on the M:N cooperative scheduler without blocking the thread that created it.

## Motivation
Certain operations, such as daemon processes, background polling loops, or long-running network listeners, are designed to outlive the function that created them. These tasks must run independently ("fire and forget"). `spawn` allows developers to schedule these background tasks efficiently without the heavy overhead of OS threads.

## Syntax

The `spawn` keyword must be followed directly by a function or method call. Optionally, it can accept a monitor channel expression enclosed in parentheses:

```nora
// Spawning a standalone function
spawn listen_for_connections(port)

// Spawning a method on an object
spawn server.start()

// Spawning with a panic monitor channel
spawn(my_chan) failing_worker()
```

## Semantics
When the Nora compiler encounters a `spawn` expression, it generates a High-Level Intermediate Representation (HIR) `Spawn` instruction. During C11 code generation, this lowers to a call to the runtime's `scheduler_spawn` function. 
1. **Allocation**: The runtime allocates a new fiber context (including a lightweight stack for Wasm/POSIX, or a native Fiber for Windows).
2. **Queuing**: The fiber is placed into the scheduler's global run queue.
3. **Execution**: The parent fiber continues executing immediately. The spawned fiber will be picked up by one of the available OS thread workers and executed cooperatively.
4. **Panic Monitoring**: If `spawn(monitor_chan)` is used, the runtime code-generation automatically wraps the fiber execution in a `setjmp` block. If the fiber panics, the panic string is converted into a Nora `str` and sent silently into the `monitor_chan` before the fiber is cleanly terminated.

## Type Rules
- `spawn` evaluates to a `void` expression. It cannot return a value directly. To return a value from a spawned fiber, a `chan[T]` (channel) must be used.
- The operand of a `spawn` must be a valid function call or method invocation.
- If the `spawn(monitor_chan)` syntax is used, the `monitor_chan` expression **must** be of type `chan[str]`.

## Lease Rules (The Topological Solver)
Because a detached `spawn` creates a fiber that runs independently and may outlive the lexical scope it was created in, the Topological Lease Solver applies extreme strictness to memory references:
1. **No Local Borrows**: You absolutely **cannot** pass read-only (`#T`) or mutable (`&T`) borrows of local variables to a detached `spawn`. Doing so violates memory safety because the local stack frame may be destroyed while the fiber is still accessing it.
2. **Ownership Transfer**: All data passed to a detached `spawn` must be passed by **value** or **ownership transfer (@)**. The fiber takes total ownership of the memory, guaranteeing it remains valid for the fiber's lifetime.

## Examples

### Detached Network Listener
```nora
fn start_server(port: i32) {
    // Starts a listener in the background indefinitely
    spawn listen_for_connections(port)
    io.PrintLn("Server started in the background.")
}
```

## Edge Cases
- **Panics**: If a detached fiber panics, the panic is caught by the fiber runtime. If a monitor channel was provided (`spawn(chan)`), the panic message is sent to the channel. If not, the panic is silently consumed (unless inside a `scope`, where it escalates to the parent). Unhandled panics in the `main` fiber terminate the process.
- **Main Termination**: If the `main` fiber finishes and returns, all currently running detached fibers are forcefully terminated.

## Errors & Diagnostics
- **Borrow Outlive Error**: If the user attempts to pass `#local_var` to a detached `spawn`, the compiler emits: `Error: Cannot borrow local variable for spawned fiber. The fiber may outlive the variable. Use a 'scope' block or pass by value (@T).`
- **Invalid Monitor Channel Type**: If the provided monitor channel is not `chan[str]`, the compiler emits: `Error: spawn monitor channel must be of type chan[str], got <type>`.

## Future Considerations
- **Generic Panic Types**: Currently `spawn(chan)` is restricted to `chan[str]`. In the future, Nora may introduce structural error enums to allow `chan[PanicReason]` for richer error matching.
