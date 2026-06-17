# Pattern Matching & Sum Types

## Overview

Nora features powerful pattern matching via the `match` expression, designed to work seamlessly with both Sum Types (`enum`) and standard structures (`struct`). The `match` expression guarantees exhaustiveness at compile time, ensuring developers handle all possible cases or explicitly use a wildcard.

## Motivation

Traditional `switch` statements in C/C++ are error-prone due to fallthrough semantics and lack of exhaustiveness checking. By using a robust `match` expression, Nora allows developers to safely destructure data types, extract inner values, and handle variants cleanly without runtime type casting.

## Syntax

### 1. Sum Types (`enum`)

Enums in Nora can define simple variants, or variants carrying associated values (Sum Types).

```nora
pub type Status = enum {
    Inactive,
    Active(val: i32),
    Pending(msg: str)
}
```

### 2. The `match` Expression

The `match` block evaluates an expression and branches based on the matched pattern. It evaluates to a value, meaning it can be assigned directly to a variable.

```nora
fn describe(s: Status) {
    match s {
        Inactive => {
            io.PrintLn("Status is Inactive")
        }
        Active(val) => {
            io.PrintLn("Status is Active with value: ${val}")
        }
        Pending(msg) => {
            io.PrintLn("Status is Pending with message: ${msg}")
        }
    }
}
```

### 3. Wildcard `_`

If you only care about specific cases, use `_` as a catch-all block.

```nora
match s {
    Active(val) => { io.PrintLn("Active") }
    _ => { io.PrintLn("Other") }
}
```

### 4. Structural Matching on Structs

`match` can destructure structs, verifying specific field values while extracting others.

```nora
type Point = struct { x: i32, y: i32 }

fn check_point(p: Point) {
    match p {
        Point { x: 0, y: 0 } => {
            io.PrintLn("Origin")
        }
        Point { x: 10, y: val } => {
            io.PrintLn("x is 10, y is ${val}")
        }
        Point { x: _, y: _ } => {
            io.PrintLn("Other point")
        }
    }
}
```

## Semantics & Type Rules

1.  **Exhaustiveness:** The compiler forces you to cover all possible variants of an `enum`. If a new variant is added to an `enum`, the compiler will flag all `match` statements missing that variant.
2.  **Shadowing/Extraction:** When matching a variant containing data (e.g., `Active(val)`), the variable `val` is locally scoped to that block.

## Lease Rules

If the variable being matched is owned (`@`), the matched bindings (if they consume data) will move ownership out of the enum/struct. Matching without consumption takes read-only leases of the inner fields.

## Examples

### Nested Structural Match
You can match deep within nested structures:
```nora
match rect {
    Rect {
        top_left: Point { x: 0, y: 0 },
        bottom_right: br
    } => {
        // Starts at origin
    }
    _ => {}
}
```

## Errors & Diagnostics

*   **Non-Exhaustive Match:** "pattern matching is non-exhaustive". Emitted when a variant or wildcard is missing.
*   **Unreachable Pattern:** If a wildcard `_` is placed before a valid pattern, the compiler will error on the subsequent unreachable branch.
