# Defer Statement

## Overview

The `defer` statement defers the execution of a function call or expression until the immediately enclosing function returns. This provides a structured, predictable way to guarantee cleanup actions without needing manual `free()` calls or complicated `try/finally` blocks.

## Motivation

In systems programming, ensuring resources (like open files, mutex locks, or socket connections) are closed during complex control flows—especially those with multiple return paths or potential errors (`?`)—is a common source of bugs. While Nora's Topological Lease Solver handles RAII for standard owned types implicitly, `defer` gives developers explicit control over when arbitrary blocks of code execute during teardown.

## Syntax

The `defer` keyword prefixes any valid expression or block of code.

```nora
fn process_file() {
    io.PrintLn("Starting processing...")
    
    defer io.PrintLn("Finished processing (cleanup)")

    io.PrintLn("Doing work...")
    // "Finished processing (cleanup)" will print AFTER this line.
}
```

## Semantics

1.  **LIFO Order:** Deferred statements are pushed onto a stack. When the function returns, they are popped off and executed in Last-In, First-Out (LIFO) order.
2.  **Function Boundary:** The `defer` statement triggers when the *enclosing function* returns, not when an inner block or `scope` ends.
3.  **Capture:** Variables used inside a `defer` block are evaluated at the time the `defer` block executes, *not* at the time it is declared.

## Examples

### Order of Execution

```nora
fn test_defer() {
    defer io.PrintLn("Deferred 1")
    defer io.PrintLn("Deferred 2")
}
// Prints:
// Deferred 2
// Deferred 1
```

### Resource Cleanup

```nora
fn complex_operation() Result[void, str] {
    var db = connect_database()
    defer db.close() // Will execute regardless of how the function exits

    var data = fetch_data(db)? // If this early returns, db.close() still runs
    
    process_data(data)
    return Ok[void, str](none)
}
```

## Errors & Diagnostics

*   **Yields No Value:** A `defer` statement does not evaluate to a value and cannot be assigned.
*   **Panic Handling:** Even if the function panics, `defer` blocks are still executed during the unwinding phase, guaranteeing resource safety.
