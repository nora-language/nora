# Lambdas & Closures

## Overview

Nora supports first-class anonymous functions (lambdas) that can capture their enclosing environment to form closures. They can be assigned to variables, passed to higher-order functions, returned from functions, and spawned into concurrent fibers.

## Motivation

Functional programming paradigms, such as `map`, `filter`, and callback-driven asynchronous logic, rely heavily on closures. Nora integrates closures while respecting the strict guarantees of the Topological Lease Solver, ensuring captured variables do not cause memory leaks or dangling pointers.

## Syntax

### 1. Basic Lambda
A lambda is declared using the `fn` keyword followed by arguments and an optional return type.

```nora
var sq = fn(x: i32) i32 {
    return x * x
};
var result = sq(6) // 36
```

### 2. Higher-Order Functions
Lambdas are seamlessly passed as arguments.

```nora
fn run_op(a: i32, b: i32, op: fn(i32, i32) i32) i32 {
    return op(a, b)
}

fn main() {
    var sum = fn(x: i32, y: i32) i32 { return x + y };
    var result = run_op(10, 20, sum) // 30
}
```

## Semantics & Variable Capturing

Closures in Nora capture variables from their outer scope. 

1.  **Value Capturing:** Primitive types (like `i32`, `bool`) are captured by value (copied).
2.  **Owned Capturing:** When a closure captures an owned reference (`@T`) or a standard heap-allocated struct, the Topological Lease Solver ensures the capture semantics match the lifecycle. 
3.  **Read-Only Leases:** Closures can capture and utilize read-only leases (`#T`).

### Frozen Variables

When a local variable is captured by a closure, that variable becomes **frozen** (immutable) in the outer scope. Any subsequent attempt to reassign or mutate the variable in the outer scope will result in a compile-time error. This prevents data races and logical inconsistencies, ensuring the closure operates on a predictable snapshot of the variable.

```nora
var x = 10
var f = fn() {
    io.PrintLn("Closure captured x = ${x}")
}

// ERROR: x cannot be modified because it is captured by the closure f
x = 20 
```

### Restricted Closures & Escape Analysis

To prevent catastrophic Use-After-Free (UAF) vulnerabilities, the compiler performs strict escape analysis. If a closure captures a local lease (`#T` or `&T`), or captures a struct that contains a lease anywhere inside its fields, it is flagged as a **Restricted Closure**.

Restricted Closures are permanently bound to the stack frame of the variables they captured. The compiler enforces strict boundaries preventing them from escaping:
1. **Cannot be returned:** You cannot `return` a restricted closure from a function.
2. **Cannot be spawned:** You cannot pass a restricted closure to `spawn` (as fibers outlive the caller).
3. **Cannot be assigned to globals:** You cannot store a restricted closure in a package-level or global variable.

```nora
fn bad_return() fn() void {
    var data = 999
    
    var closure = fn() void {
        var local_val: #i32 = #data
    }
    
    // ERROR: closure captures a local lease and cannot escape via return!
    return closure 
}
```

```nora
var limit = 5
var check_limit = fn(val: i32) bool {
    // limit is captured from the outer scope
    return val <= limit 
};
```

## Lease Rules

Because Nora avoids hidden garbage collection, returning a closure that captures a stack variable by reference is forbidden. Closures typically clone their captured environments dynamically into a heap-allocated context block when they are boxed into `fn(...)` interface wrappers under the hood.

*   If a closure captures an owned object (`@Vector[i32]`), that object is moved into the closure's hidden state. The closure now physically owns the object. 
*   When the closure is ultimately dropped by the lease solver, it automatically triggers the `drop()` of its captured owned objects.

## Examples

### Closures in Concurrency

Closures are extremely powerful when combined with the `spawn` keyword for concurrent execution.

```nora
var factor = 3
var task = fn() i32 {
    var local_val = 14
    return local_val * factor // Captures 'factor'
};

var done_ch = make(chan[i32], 1)
scope {
    spawn concurrent_worker(task, #done_ch)
    var result_val = <-done_ch // 42
}
```

## Errors & Diagnostics

*   **Capture Move Error:** If you attempt to capture an already-moved variable in a closure, the compiler will yield a "use-after-move" error.
