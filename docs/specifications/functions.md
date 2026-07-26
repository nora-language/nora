# Functions

## Overview

Functions are the fundamental building blocks of execution in Nora. Declared using the `fn` keyword, they enforce strong static typing, explicit ownership boundaries for parameters, and support early control-flow termination via `return`.

## Syntax

A function is declared with `fn`, an optional visibility modifier `pub`, a name, a parameter list, and an optional return type.

```nora
pub fn calculate_area(width: i32, height: i32) i32 {
    return width * height
}
```

If a function does not return a value, the return type is omitted (implicitly `void`).

```nora
fn print_greeting(name: str) {
    io.PrintLn("Hello, ${name}")
}
```

## Parameter Leasing & Ownership

Nora's memory safety relies heavily on how parameters are passed into functions. By default, primitive types (`i32`, `bool`) are passed by value (copied). However, for complex types (structs, collections), you must explicitly declare the lease rules.

1.  **Owned (`@T`):** The function takes ownership of the data. The caller loses access.
2.  **Read-Only Lease (`#T`):** The function borrows the data without mutating it. The caller retains access.
3.  **Mutable Lease (`&T`):** The function borrows the data exclusively to mutate it.

```nora
pub fn consume_data(data: @Vector[i32]) { ... }
pub fn read_data(data: #Vector[i32]) { ... }
pub fn mutate_data(data: &Vector[i32]) { ... }
```

### Returning Leases
Functions can also explicitly return leases (`&T` or `#T`). This is heavily utilized when accessing internal elements of collections or data structures without forcing a value copy. The Topological Lease Solver will track the returned lease and map its dependency back to the parent provider.

```nora
pub fn get_element_mut(vec: &Vector[i32], index: i32) &i32 {
    return &vec.data[index]
}
```

## Methods (Receiver Functions)

Nora supports attaching functions to structs (methods) by defining a "receiver" before the function name.

```nora
type Circle = struct { radius: f64 }

// 'self' is conventionally used, though any name works.
pub fn (self: #Circle) Area() f64 {
    return 3.14 * self.radius * self.radius
}
```

## Overloading

Nora supports standard **Function Overloading** and **Method Overloading** based on parameter types. You can define multiple functions or methods with the exact same name, as long as their parameter signatures are strictly distinct.

### Function Overloading

```nora
pub fn print_val(val: i32) {
    io.PrintLn("Integer: ${val}")
}

pub fn print_val(val: f64) {
    io.PrintLn("Float: ${val}")
}

pub fn main() {
    print_val(10)     // Matches the i32 overload
    print_val(10.5)   // Matches the f64 overload
}
```

### Method Overloading

```nora
type Printer = struct {}

pub fn (p: #Printer) print(val: i32) {
    io.PrintLn("Integer: ${val}")
}

pub fn (p: #Printer) print(val: str) {
    io.PrintLn("String: ${val}")
}
```

### Overload Resolution Rules
- The Semantic Analyzer groups overloaded functions into a `SymOverloadGroup`.
- When a call is encountered, the compiler resolves the best matching function by strictly comparing the types of the arguments provided against the parameters of the available overloads.
- If multiple overloads ambiguously match or no overload matches, a semantic error is emitted.
- At the code generation level, the C backend uniquely mangles overloaded functions by appending a type-based hash suffix to their names to ensure binary-level uniqueness.

## Return Values

The `return` keyword is used to exit a function and optionally yield a value back to the caller.

```nora
fn get_status() i32 {
    if condition {
        return 1
    }
    return 0
}
```

Nora does not require explicit `return` at the end of `void` functions. However, if a function declares a return type, the compiler mandates that all possible control flow paths end in a valid `return` statement (or a `panic`).

## Errors & Diagnostics

*   **Missing Return:** If a function declares a return type but a branch fails to return a value, the compiler errors.
*   **Type Mismatch:** Returning a value that does not match the declared return type yields an error.
