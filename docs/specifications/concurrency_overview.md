# Concurrency in Nora: Comparison and Overview

## Title & Overview
Nora provides a powerful, memory-safe M:N cooperative scheduling model with zero-cost abstractions. To handle different concurrency use-cases without incurring the heavy overhead of OS threads, Nora splits concurrency into four fundamental primitives: `spawn`, `scope`, `parallel`, and `chan[T]`. 

This document serves to explain the explicit differences between them and when to use each.

## 1. Execution Control vs. Communication

The primitives are divided into two categories:
*   **Execution Control** (`spawn`, `scope`, `parallel`): These dictate *how* and *when* fibers run, and how the parent thread waits for them (if at all).
*   **Communication** (`chan[T]`): This dictates *how* running fibers talk to each other and share data safely without locks.

---

## 2. Execution Control Comparison

### A. `spawn` (The Raw Primitive)
- **Behavior**: "Fire-and-forget". It tosses a function onto the scheduler's queue and instantly returns to the caller. The parent thread does not wait.
- **Lease Rules**: Extremely strict. Cannot borrow local variables (`#T` or `&T`). Demands complete ownership (`@T`) of data because the spawned fiber might outlive the parent function.
- **Best For**: Daemon tasks, background listeners, or infinite loops (e.g., listening for incoming HTTP requests on a server).
```nora
spawn listen_for_connections() // Runs forever in the background
```

### B. `scope` (The Structural Manager)
- **Behavior**: "Managed spanning". It defines a block where multiple `spawn` calls can be made. The closing brace `}` acts as an implicit WaitGroup, forcing the parent thread to wait until all inner spawns are completed.
- **Lease Rules**: Relaxed. Because the parent thread is guaranteed to wait, fibers spawned inside a `scope` are permitted to borrow (`#T` or `&T`) local variables from the parent without violating memory safety.
- **Best For**: Dynamic fan-out tasks where you need to spawn an arbitrary number of fibers in a loop and wait for them all to finish.
```nora
scope {
    for var i = 0; i < 5; i++ {
        spawn worker(i, #shared_data) 
    }
} // Blocks here until all 5 workers finish
```

### C. `parallel` (The CPU-Bound Optimizer)
- **Behavior**: "Data-parallelism block". It takes every top-level statement inside its block and aggressively distributes them across physical CPU cores simultaneously. It implicitly blocks at the end. Note that you do **not** use the `spawn` keyword inside a `parallel` block.
- **Lease Rules**: Relaxed. Similar to `scope`, it allows safe borrowing (`#T` or `&T`). However, if two statements try to mutably borrow (`&T`) the exact same variable, the compiler will instantly throw a Data Race error.
- **Best For**: Fixed sequences of heavy mathematical or CPU-bound computations that can be evaluated entirely independently.
```nora
parallel {
    process_left_half(#data)  // Runs on Core A
    process_right_half(#data) // Runs on Core B
} // Blocks here until both halves complete
```

---

## 3. Communication: Channels (`chan[T]`)

While the previous three keywords control execution, **channels** are how you wire them together. 

- **Behavior**: A zero-copy data pipe. It allows one fiber to send data ownership directly into another fiber.
- **Synchronization**: Channels inherently synchronize fibers. If a fiber tries to read from an empty channel, it will cooperatively suspend itself and go to sleep. When another fiber sends data into that channel, the sleeping receiver is immediately woken up by the scheduler.
- **Best For**: Streaming data between fibers, producer-consumer patterns, or returning results from a detached `spawn` fiber back to the main thread.
```nora
var pipe = alloc make(chan[i32], 5)

scope {
    spawn producer(#pipe)
    spawn consumer(#pipe)
}
```

## Summary Checklist

| Feature | Blocks Parent Thread? | Borrows Allowed? | When to use? |
| :--- | :--- | :--- | :--- |
| **`spawn`** | No | No (Ownership `@` only) | Background daemons, independent tasks |
| **`scope`** | Yes (At the `}` brace) | Yes (`#`, `&`) | Managing multiple dynamic `spawn` calls |
| **`parallel`** | Yes (At the `}` brace) | Yes (`#`, `&`) | Heavy, isolated, CPU-bound tasks |
| **`chan[T]`** | Yes (If full/empty) | Transfers Ownership (`@`) | Sending data between fibers safely |
