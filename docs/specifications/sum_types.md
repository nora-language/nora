# Sum Types (`enum`)

## Overview

Nora uses the `enum` keyword to define Algebraic Data Types (ADTs), specifically Sum Types. Unlike traditional C-enums that merely assign names to integers, Nora's enums can carry associated data payloads within their variants.

## Motivation

Sum types allow developers to model states that are mutually exclusive but carry different data, such as a state machine, network response, or error result, in a completely type-safe and memory-safe way.

## Syntax

An `enum` is declared as a `type` containing a comma-separated list of variants. Variants can optionally define a tuple or struct-like payload.

```nora
pub type Status = enum {
    Inactive,
    Active(val: i32),
    Pending(msg: str)
}
```

### Instantiation

To instantiate an enum, use the variant name directly (variants are implicitly scoped to the package, or accessible via dot notation if imported).

```nora
var s1 = Inactive
var s2 = Active(42)
var s3 = Pending("Wait for it")
```

## Memory Layout & Semantics

1.  **Tagged Union:** Under the hood, an `enum` is compiled to a C11 `struct` containing an integer tag (the discriminator) and a `union` of all the variant payloads.
2.  **Size:** The size of an `enum` is the size of the tag plus the size of its largest variant payload, plus any necessary alignment padding.
3.  **Pattern Matching:** The only safe way to extract data from an `enum` variant is by using the `match` or `if match` expressions. This prevents unsafe arbitrary union access and guarantees exhaustiveness.

```nora
match s2 {
    Active(val) => { io.PrintLn("Value: ${val}") }
    _ => {}
}
```

## Generic Sum Types

Sum types can be generic. The most common examples in Nora's standard library are `Option[T]` and `Result[T, E]`.

```nora
pub type Option[T] = enum {
    Some(val: T),
    None
}
```

## Errors & Diagnostics

*   **Variant Not Found:** Using a variant name that is not defined in the enum yields a compiler error.
*   **Missing Payload:** Attempting to instantiate a variant that requires a payload without providing one yields a compiler error.
