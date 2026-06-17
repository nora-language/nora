package main

import "io"

// 1. Function type as a field
type Processor[T] = struct {
    handler: fn(#T) i32
}

fn apply[T](p: #Processor[T], val: #T) i32 {
    return p.handler(val)
}

// 2. Concrete implementation
fn my_handler(val: #i32) i32 {
    return val * 2
}

fn main() i32 {
    var p = Processor[i32] {
        handler: my_handler
    }

    var x = 21
    var res = apply[i32](p, x)
    io.PrintLn("Result: ${res}") // Should be 42

    // 3. Anonymous function type variable
    var f: fn(#i32) i32 = my_handler
    var y = 10
    io.PrintLn("Direct call: ${f(y)}") // Should be 20

    return 0
}
