# Variables & Definite Initialization

## Overview

Variables in Nora are declared using the `var` keyword. To prevent subtle logic bugs and undefined behavior caused by uninitialized memory (a notorious problem in C), Nora employs a strict **Definite Initialization** static analysis pass.

## Syntax & Type Inference

A variable can be explicitly typed or implicitly typed via type inference.

### 1. Type Inference
If you provide an initial value, Nora infers the type automatically.

```nora
var x = 10        // Inferred as i32
var name = "Nora" // Inferred as str
var is_active = true // Inferred as bool
```

### 2. Explicit Typing
You can explicitly define the type, which is necessary if you wish to declare a variable without immediately assigning it a value.

```nora
var x: i32 = 10
var y: i64        // Declared, but uninitialized
```

## Definite Initialization

Nora guarantees that a variable is never read before it has been definitively initialized. The semantic analyzer traces every possible control flow branch (e.g., `if`, `while`, `match`) to verify initialization.

### Valid Initialization

```nora
fn process(condition: bool) {
    var result: i32

    if condition {
        result = 10
    } else {
        result = 20
    }

    // This is SAFE. result is initialized in all branches.
    io.PrintLn("Result is ${result}")
}
```

### Invalid Initialization (Compile Error)

```nora
fn calculate(condition: bool) {
    var val: i32

    if condition {
        val = 100
    }

    // ERROR: val used without being definitively initialized!
    // If condition was false, val holds garbage data.
    io.PrintLn("Value is ${val}") 
}
```

## Reassignment and Leases

*   **Mutating:** Reassigning a variable is allowed by default. 
*   **Leak Prevention:** If the variable holds an owned (`@`) resource (like a heap allocation), the topological lease solver will automatically insert a `PreDrop` instruction before the reassignment occurs to free the old memory and prevent a leak.

```nora
var buffer = alloc Vector[i32]()
buffer = alloc Vector[i32]() // The first buffer is automatically dropped here
```

## Errors & Diagnostics

*   **Use of uninitialized variable:** Triggers when the static analysis detects a read operation on a variable that lacks definite initialization along any control flow path.
