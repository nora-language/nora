# Structured Concurrency: The `scope` Keyword

## Title & Overview
The `scope` keyword in Nora introduces structured concurrency. It defines a lexical boundary for asynchronous fibers, guaranteeing that all fibers spawned within the block will complete their execution before the block exits. By implicitly injecting a `WaitGroup`, the `scope` keyword prevents orphaned background tasks and enables safe, statically-verified borrowing of local variables into concurrent fibers.

## Motivation
In traditional concurrent programming (like Go's goroutines), spawned tasks are "fire and forget". This leads to two major problems:
1. **Orphaned Tasks**: If a parent function returns early or crashes, the background tasks continue running invisibly, leaking resources or causing erratic behavior.
2. **Dangling References**: A spawned task cannot safely borrow a reference (e.g., `#data` or `&data`) to a variable on the parent's stack because the parent might return and destroy the stack frame before the task finishes. 

Nora solves both problems natively through the `scope` keyword. By guaranteeing that the parent thread blocks until all children finish, Nora's Topological Lease Solver can safely allow fibers to borrow references to data that outlives the scope block.

## Syntax

A `scope` is declared using the `scope` keyword followed by a block of statements:

```nora
scope {
    spawn do_work(1)
    spawn do_work(2)
}
```

Nested scopes are also fully supported:
```nora
scope {
    spawn worker_a()
    scope {
        spawn worker_b()
    }
}
```

## Semantics

Behind the scenes, the Nora compiler performs AST lowering to automatically inject synchronization primitives. 

1. **Initialization**: When entering a `scope` block, the compiler implicitly allocates a WaitGroup (e.g., `_scope_wg_1 = nr_sync_waitgroup_create()`).
2. **Registration**: For every `spawn` expression executed inside the `scope`, the compiler automatically increments the WaitGroup counter (`nr_sync_waitgroup_add(..., 1)`).
3. **Execution**: The spawned fiber's wrapper function is compiled to automatically call `nr_sync_waitgroup_done(...)` when the fiber finishes (either normally or via an early return).
4. **Completion**: At the closing brace `}` of the `scope`, the compiler injects a blocking wait call (`nr_sync_waitgroup_wait(...)`). The parent thread executing the scope will block at this point until the WaitGroup counter reaches zero.
5. **Cleanup**: Once all fibers complete, the waitgroup is destroyed and execution continues past the scope block.

## Type Rules
- `scope` is a block statement, not an expression. It does not evaluate to a value.
- The `spawn` keyword inside a `scope` behaves identically to a normal `spawn`, but with altered topological lease requirements.

## Lease Rules (The Topological Solver)
Normally, a `spawn` requires the target closure or function arguments to take **ownership** (`@T`) of variables, because the fiber's lifetime is untracked and could outlive the parent. 

When `spawn` is used **inside** a `scope` block:
1. The Topological Lease Solver temporarily relaxes lifetime bounds for variables declared *outside* and *before* the `scope`.
2. Fibers spawned inside the scope are safely permitted to take Read-Only Borrows (`#T`) or Mutable Borrows (`&T`) of those outer variables.
3. The solver enforces that the outer variables cannot be moved, consumed, or destroyed until the `scope` block has fully closed.

## Examples

### Safe Borrowing across Fibers

```nora
fn process_parallel() {
    var pool = alloc SharedPool { ... }

    // Using scope allows fibers to safely borrow `#pool`
    scope {
        spawn worker(1, #pool)
        spawn worker(2, #pool)
        spawn worker(3, #pool)
    }

    // This line is only reached when all 3 workers have completely finished.
    io.PrintLn("All workers are done processing!")
}
```

## Edge Cases
- **Empty Scopes**: A `scope { }` with no `spawn` statements compiles down to a no-op with zero runtime overhead.
- **Dynamic Spawns**: If `spawn` is used inside a `while` loop within a `scope`, the WaitGroup correctly increments for every iteration, ensuring all dynamically spawned fibers are awaited.
- **Early Returns**: Returning from inside a `scope` block will automatically trigger the implicit `wg.Wait()` before the return instruction is actually executed, ensuring children are not orphaned.

## Errors & Diagnostics
- **Orphaned Borrow Error**: Attempting to pass a borrow (`#T` or `&T`) to a `spawn` *outside* of a `scope` will result in a compiler diagnostic: `Error: Cannot borrow local variable for spawned fiber. The fiber may outlive the variable. Use a 'scope' block or pass by value (@T)`.

## Future Considerations
- **Scope Cancellation**: Future iterations of `scope` may automatically inject cancellation contexts, allowing the parent to terminate all children fibers if one of the children panics or fails.
- **Error Propagation**: Allowing a `scope` to return an aggregated `Result` if one or more spawned fibers return an error state.
