package main

import "io"

// This function will be callable from C as 'NORA_add'
export fn NORA_add(a: i32, b: i32) i32 {
    io.PrintLn("Nora: Adding ${a} and ${b}")
    return a + b
}

fn main() i32 {
    io.PrintLn("Main running")
    return 0
}
