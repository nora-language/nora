# Structs

## Overview

A `struct` is the primary way to define complex custom data types in Nora. It allows you to group related fields of varying types into a single contiguous block of memory.

## Syntax

### 1. Definition

A struct is defined using the `type` keyword, assigning a name to a `struct { ... }` block.

```nora
pub type Point = struct {
    x: i32,
    y: i32
}
```

You can control field visibility individually using the `pub` keyword. By default, fields are private to the module.

```nora
pub type User = struct {
    pub id: i32,
    password_hash: str
}
```

### 2. Instantiation

To create a new instance of a struct, use the struct name followed by a block of field initializations.

```nora
var p = Point {
    x: 10,
    y: 20
}
```

If a struct is allocated on the heap, you must use the `alloc` keyword.

```nora
var hp = alloc Point {
    x: 5,
    y: 5
}
```

### 3. Field Access

Fields are accessed using dot notation (`.`).

```nora
io.PrintLn("X coordinate is ${p.x}")
p.x = 15
```

## Semantics & Memory Layout

1.  **C-Compatible Layout:** By default, Nora compiles structs down to C11 `struct` definitions. The memory layout is sequential, subject to standard C alignment rules (padding may be inserted by the C compiler).
2.  **Move Semantics (Default):** By default, stack-allocated structs act as owned, linear types. Assigning a struct to a new variable or passing it to a function transfers ownership (moves it). Using the original variable after a move yields a `use of moved value` compiler error, unless passed via a lease (`#`, `&`).
3.  **Copy Semantics:** To make a struct behave like a primitive type (where assignments perform implicit copies instead of moves), annotate it with the `[copyable]` attribute.

## Errors & Diagnostics

*   **Missing Fields:** When instantiating a struct, failing to provide an initialization value for every field yields a compile-time error. (Nora does not zero-initialize structs implicitly; Definite Initialization rules apply to struct instantiation).
*   **Visibility Error:** Attempting to read or write a private field from outside the package yields a compile error.
