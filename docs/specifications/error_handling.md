# Error Handling

## Overview

Nora handles errors purely as values, utilizing algebraic data types (specifically the `Result` sum type) to represent success and failure states. There are no invisible exceptions thrown or caught in standard control flows.

## Motivation

Hidden control flow (exceptions) makes it difficult to reason about what parts of a program might fail and what the state of memory is after a failure. By making errors explicit values returned from functions, developers are forced to acknowledge and handle them, leading to robust and predictable software.

## Syntax

### 1. The Result Type

The built-in `Result[T, E]` generic type is a sum type representing either success (`Ok(T)`) or failure (`Err(E)`).

```nora
pub fn divide(a: i32, b: i32) Result[i32, str] {
    if (b == 0) {
        return Err[i32, str]("division by zero")
    }
    return Ok[i32, str](a / b)
}
```

### 2. The Try Operator (`?`)

The `?` suffix operator provides ergonomic error propagation. When appended to an expression that returns a `Result`, it automatically unwraps the `Ok` value if successful. If it evaluates to an `Err`, the enclosing function immediately returns the error.

```nora
pub fn calculate_complex(a: i32, b: i32, c: i32, d: i32) Result[i32, str] {
    // If divide() returns an Err, calculate_complex returns it immediately.
    var x = divide(a, b)? + divide(c, d)?
    return Ok[i32, str](x)
}
```

## Semantics & Type Rules

1.  **Exhaustiveness:** You can use pattern matching (`match`) to explicitly handle both `Ok` and `Err` variants of a `Result`.
2.  **Try Operator Requirements:** The `?` operator can only be used inside functions that themselves return a `Result` type compatible with the error being propagated.
3.  **Panics:** For unrecoverable errors (e.g., out-of-bounds array access), Nora provides `panic`. Unlike exceptions, panics crash the fiber/thread and are not meant to be caught as standard control flow.

## Examples

### Handling Errors with Match

```nora
fn main() {
    var res = divide(10, 0)
    match res {
        Ok(val) => {
            io.PrintLn("Result: ${val}")
        }
        Err(msg) => {
            io.PrintLn("Error: ${msg}")
        }
    }
}
```

## Errors & Diagnostics

*   **Incompatible Error Return:** Attempting to use `?` in a function that returns `void` or a different error type than the expression will result in a compile-time type checking error.
*   **Unused Result:** Ignoring a returned `Result` type without explicitly discarding it triggers a compiler warning, as it implies an unhandled potential failure.

## Future Considerations

*   **Error Interface:** Introduction of a standard `Error` interface/protocol to allow for richer, composable error payloads beyond standard strings or enums.
