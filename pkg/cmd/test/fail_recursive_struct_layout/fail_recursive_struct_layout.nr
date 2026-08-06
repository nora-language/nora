package main

import "io"

// ERROR: Inline self-referential struct without pointer wrapper is an infinite size type
pub type RecursiveNode = struct {
    val: i32,
    next: RecursiveNode
}

fn main() i32 {
    io.PrintLn("This should fail semantic analysis")
    return 0
}
