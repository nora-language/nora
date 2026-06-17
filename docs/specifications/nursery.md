# Structured Concurrency: Nursery

## Overview

While the `scope` block natively provides static structured concurrency (spawning a predetermined number of fibers within a lexical block and waiting for them), Nora provides the `nursery` standard library module for **Dynamic Structured Concurrency**.

## Motivation

When writing server software, you often need to spawn an arbitrary, unbound number of tasks (e.g., handling incoming connections in a loop) and aggregate their results or failures. A `scope` block cannot easily manage dynamic spawns across function boundaries or handle aggregated error reporting dynamically. A `nursery` solves this.

## The Nursery Module (`import "nursery"`)

A nursery is a struct representing a dynamic collection of spawned fibers. The parent thread explicitly waits for the nursery to empty, and the nursery transparently aggregates any errors or panics that occur within its child fibers.

### Syntax & Usage

```nora
import "nursery"

fn worker(n: #nursery.Context, id: i32) {
    // Do work...
    if error_occurred {
        n.done("Connection timed out")
    } else {
        n.done("") // Success
    }
}

fn main() {
    var n = nursery.New()

    var i = 0
    while i < 3 {
        n.start_task() // Register a new task
        spawn(n.completion_chan) worker(#n, i)
        i++
    }

    // Wait for all tasks to call n.done()
    var res = n.wait()

    match res {
        Ok(_) => { io.PrintLn("All workers succeeded!") }
        Err(err) => { io.PrintLn("A worker failed: " + err) }
    }
}
```

## Semantics

1.  **Registration (`start_task`):** Increments the internal counter of expected tasks.
2.  **Completion (`done`):** A worker signals it is finished. If an error string is provided, the nursery captures it.
3.  **Synchronization (`wait`):** The parent thread yields cooperatively until the counter reaches zero. If any task called `done` with an error, or if any task `panic`ed, `wait()` will return an `Err` encapsulating the failure.

## Error & Panic Propagation

One of the most powerful features of the nursery is panic propagation. If a spawned fiber panics, the panic is caught by the fiber runtime and sent through the nursery's completion channel. The `n.wait()` call translates this panic into an aggregated `Err` return value, preventing a single background worker from crashing the entire application undetected.
