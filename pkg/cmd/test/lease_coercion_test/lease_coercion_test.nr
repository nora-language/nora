package main

import "io"

fn test_coercion(a: #i32) i32 {
    return a
}

fn main() {
    var val: i32 = 42
    var b_val = test_coercion(#val)
    io.PrintLn("Coercion result: ${b_val}")
}
