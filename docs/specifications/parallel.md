# Multi-Core Parallelism: The `parallel` Block

## Title & Overview
The `parallel` keyword in Nora provides highly optimized, CPU-bound parallelism. It allows developers to specify a block of statements that will be automatically transformed into isolated fibers, aggressively distributed across all available OS threads, and executed simultaneously. The block acts as an implicit synchronization barrier, ensuring the parent thread waits for all parallel operations to complete.

## Motivation
When performing heavy data processing, developers typically have to manually instantiate `WaitGroup`s, use `spawn` multiple times, and handle synchronization. `parallel` provides "syntactic sugar" for extreme data-parallelism, allowing developers to write concurrent numerical or array processing code as if it were sequential, while the compiler handles the complex distribution physics.

## Syntax

The `parallel` keyword is followed by a block of statements. **Note that `spawn` is not used inside the block.** Every statement inside the block is implicitly treated as a parallel operation.

```nora
parallel {
    process_chunk_a(data_a)
    process_chunk_b(data_b)
    process_chunk_c(data_c)
}
// Automatically blocks here until A, B, and C complete
```

## Semantics
When the compiler encounters a `parallel` expression:
1. **Implicit Wrapper Generation**: The compiler takes every individual statement inside the block and lifts it into a hidden, automatically generated wrapper function (e.g., `__parallel_wrapper_X`).
2. **Implicit WaitGroup**: An implicit WaitGroup is created, identical to the `scope` block.
3. **Scatter & Gather**: The runtime schedules every wrapper function to the global fiber queue simultaneously (scatter). The scheduler's M:N OS thread workers immediately pick up the fibers and execute them physically in parallel.
4. **Barrier**: The parent thread invokes a `Wait()` on the WaitGroup (gather), halting execution until the final parallel statement finishes.

## Type Rules
- `parallel` is a block statement and does not evaluate to a value.
- Statements inside a `parallel` block must be valid expressions or function calls.
- `return` statements are strictly prohibited inside a `parallel` block, as returning from a dynamically generated parallel fiber cannot logically return from the parent function.

## Lease Rules (The Topological Solver)
The Topological Lease Solver treats `parallel` identical to `scope` when it comes to memory safety:
1. **Safe Borrowing Allowed**: Because the `parallel` block acts as a hard synchronization barrier (joining all operations at the closing brace `}`), the parallel fibers are **guaranteed** to finish before the parent function continues.
2. **Read/Write Borrows**: Therefore, statements inside a `parallel` block are permitted to borrow (`#T`) memory that lives outside the block.
3. **Mutable Borrowing Exception**: If multiple parallel statements attempt to take a mutable borrow (`&T`) of the *same* variable, the Topological Solver will emit a **Data Race** error, ensuring memory safety across threads.

## Examples

### Heavy Matrix Multiplication
```nora
fn compute_matrices(matrix_a: #Matrix, matrix_b: #Matrix) {
    // Both halves are computed simultaneously on separate cores
    parallel {
        compute_top_half(#matrix_a)
        compute_bottom_half(#matrix_b)
    }
    
    io.PrintLn("Both halves finished computing!")
}
```

## Edge Cases
- **Overhead**: Spawning parallel fibers has a small but non-zero cost (fiber allocation). If the operations inside the `parallel` block are extremely trivial (e.g., `x = 1 + 1`), the scheduling overhead may outweigh the parallelism benefits.
- **Looping**: Currently, `parallel` operates on lexical statements. To run loops in parallel, users must chunk the data and call a processing function for each chunk inside the block.

## Errors & Diagnostics
- **Parallel Data Race Error**: If the solver detects that two parallel statements mutate the same variable, it halts compilation.

## Future Considerations
- **Parallel For-Loops**: Expanding the syntax to support `parallel for`, allowing arrays to be implicitly chunked and processed across cores without manual subdivision.
