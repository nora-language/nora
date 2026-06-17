# Standard Prelude

## Overview

The Nora Standard Prelude (`std/prelude.nr`) contains the most fundamental generic types and algebraic data types (ADTs) used across all Nora codebases. The compiler automatically injects an `import "prelude"` statement into every compiled file, meaning these types and methods are globally available without explicit imports.

## `Option[T]`

The `Option[T]` sum type represents the encapsulation of an optional value: every `Option` is either `Some` and contains a value, or `None`, and does not. This completely eliminates null-pointer exceptions in safe Nora code.

```nora
// Implicitly available in all files
pub type Option[T] = enum {
    Some(T),
    None
}
```

### Methods
*   `IsSome[T]() bool`: Returns `true` if the option is a `Some` value.
*   `IsNone[T]() bool`: Returns `true` if the option is a `None` value.
*   `Unwrap[T]() T`: Returns the inner `T` value. **Panics** if the value is a `None`.
*   `UnwrapOr[T](def_val: T) T`: Returns the contained `Some` value or a provided default `def_val`.

## `Result[T, E]`

The `Result[T, E]` sum type is the foundation of Nora's error handling. It represents either success (`Ok`) containing a value of type `T`, or failure (`Err`) containing an error of type `E`.

```nora
// Implicitly available in all files
pub type Result[T, E] = enum {
    Ok(T),
    Err(E)
}
```

### Methods
*   `IsOk[T, E]() bool`: Returns `true` if the result is `Ok`.
*   `IsErr[T, E]() bool`: Returns `true` if the result is `Err`.
*   `Unwrap[T, E]() T`: Returns the contained `Ok` value. **Panics** if the value is an `Err`.
*   `UnwrapOr[T, E](def_val: T) T`: Returns the contained `Ok` value or a provided default `def_val`.
*   `UnwrapErr[T, E]() E`: Returns the contained `Err` error value. **Panics** if the value is an `Ok`.

## `Range`

The `Range` struct represents a bounded, lazily evaluated sequence of integers. It is automatically instantiated under the hood when a developer uses the Range Operator (`..`) within a `for-in` loop, or can be used explicitly as an iterator.

```nora
pub type Range = struct {
    start: i32
    end: i32
    current: i32
}
```

### Methods
*   `Next() Option[i32]`: Advances the range iterator and returns the next integer. Returns `None[i32]` when the sequence reaches the `end` boundary.
