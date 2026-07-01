# Generics Specification

## Overview

Nora provides a comprehensive generics system designed to allow developers to write reusable, type-safe data structures and functions without sacrificing runtime performance. Unlike C++ templates or Rust generics which heavily rely on monomorphization for every distinct type combination (leading to binary bloat and slow compile times), Nora employs a hybrid approach: **Type-Erased Shared Monomorphization**.

## Motivation

Writing reusable algorithms and data structures requires a robust generics system. However, traditional monomorphization creates a unique copy of a generic function or struct for every instantiated type parameter. Nora solves the resulting binary bloat by sharing compiled instances for all "pointer-like" types under the hood, reserving unique monomorphization only for types with different value layout configurations (e.g., primitive integers).

## Syntax

Generics in Nora use square brackets `[T]` instead of angle brackets (`<T>`), improving parser efficiency and avoiding ambiguity with comparison operators.

### 1. Generic Structs

```nora
pub type Vector[T] = struct {
    data: @T[],
    length: i32,
    capacity: i32
}
```

### 2. Generic Functions

```nora
pub fn CreateBox[T](initial: T) @Box[T] {
    return alloc Box[T] {
        value: initial
    }
}
```

### 3. Generic Methods

Methods on generic types must declare the type parameter both in the receiver and on the function name if needed.

```nora
pub fn (v: &Vector[T]) Push[T](val: T) {
    // ... Implementation ...
}
```

### 4. Generic Constraints

You can restrict generic types to only those that implement a specific interface by using the syntax `[T: InterfaceName]`. This ensures type safety and method availability at compile time without relying on runtime dynamic dispatch.

#### Interface Constraints

```nora
pub type Printable = interface {
    fn custom_print() void
}

pub fn print_it[T: Printable](x: T) {
    x.custom_print() // Safe: T is guaranteed to have custom_print()
}
```

#### The `Copy` Protocol

By default, the Nora compiler considers all generic variables `T` as owned, linear types. This means that passing a generic `T` variable to another function or performing arithmetic operations on it will consume (move) the value.

To explicitly declare that a generic type `T` is trivially copyable (e.g., primitives like `f64` or `i32`) and should not be consumed by value, use the built-in `Copy` protocol constraint:

```nora
pub fn SupportArithmetic[T: Copy](a: T, b: T) T {
    var offset = a - b
    var offset2 = a + b // Safe: 'a' and 'b' are copied, not moved
    return offset2
}
```

Attempting to pass an owned type (like a `str` or an `@Vector`) into a `[T: Copy]` constrained function will trigger a semantic error.
## Semantics & Type-Erased Monomorphization

At compile time, Nora's `pkg/codegen` evaluates generic instantiations. 

If a generic struct, sum type, or function is instantiated with *only* pointer-like arguments:
- Pointers (`ptr`)
- Strings (`str`)
- Channels (`chan`)
- Other generic pointer types

Nora merges them into a single, shared, type-erased implementation suffixing the generated C functions with `_ptr`. This significantly reduces binary size.
For primitive types with unique memory layouts (`i32`, `f64`, etc.), a distinct concrete monomorphized variant is instantiated.

## Type Rules

1.  **Instantiation:** A generic type or function must be instantiated with concrete types explicitly when the compiler cannot infer them.
2.  **Inference:** In many contexts, such as function arguments, Nora can infer the type parameters `T` from the passed arguments.
    ```nora
    // Type inferred as i32
    var b = CreateBox(42) 
    ```

## Lease Rules

Generic types correctly respect Nora's Topological Lease Solver. If a generic structure contains an owned element (`@T`), dropping the generic structure will recursively trigger the `drop()` method of the underlying concrete type `T` at compile time.

## Examples

### Complete Generic Queue

```nora
pub type Node[T] = struct {
    value: T,
    next: @Node[T]
}

pub type Queue[T] = struct {
    head: @Node[T]
}

pub fn (q: &Queue[T]) Enqueue[T](val: T) {
    var new_node = alloc Node[T] {
        value: val,
        next: none
    }
    // ...
}
```

## Errors & Diagnostics

*   **Type Mismatch in Instantiation:** Attempting to assign `Box[str]` to `Box[i32]` will result in a hard compilation error.
*   **Method Signature Mismatch:** The generic arguments in the method receiver must perfectly match the generic parameters of the struct definition.
*   **Constraint Violation:** Passing a type to a generic function that does not implement the required constraint interface will trigger a compile-time error.
