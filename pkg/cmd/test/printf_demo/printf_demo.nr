package main

import "io"

// Testing ... variadic support
fn main() i32 {
    var x = 10
    var y = 20
    var z = x + y

    io.PrintLn("Result: ${x} + ${y} = ${z}")
    io.PrintLn("Sum: ${z}")

    return 0
}
