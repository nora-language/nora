package main

import "io"

// A simple copyable struct representing coordinates
[copyable]
pub type Point = struct {
    x: f64,
    y: f64
}

fn main() i32 {
    // 1. STACK ALLOCATION (Value Type)
    // Here we are creating a Point value directly on the stack.
    // The type of 'p1' is: Point
    var p1 = Point{ x: 10.0, y: 20.0 }
    
    // Because Point is [copyable], this performs an implicit bitwise copy.
    // Ownership is NOT transferred. Both 'p1' and 'p2' are completely valid.
    var p2 = p1 
    io.PrintLn("Stack: p1.x = ${p1.x}, p2.x = ${p2.x}")


    // 2. HEAP ALLOCATION (Pointer Type)
    // By using 'alloc', we allocate a Point on the heap and get an owned pointer.
    // The type of 'ptr_a' is: @Point (a linear Move type)
    var ptr_a = alloc Point{ x: 100.0, y: 200.0 }

    // If we try to do `var ptr_b = ptr_a`, the compiler would MOVE the pointer.
    // The pointer itself is NEVER copyable because it owns the heap memory.
    
    // We can safely modify the heap allocation using auto-dereferencing field access
    ptr_a.x = 999.0

    // Both the stack values and the heap pointer are valid!
    io.PrintLn("Heap Pointer: ptr_a.x = ${ptr_a.x}")
    
    // At the end of the scope, 'ptr_a' is automatically dropped and frees the heap memory.

    return 0
}
