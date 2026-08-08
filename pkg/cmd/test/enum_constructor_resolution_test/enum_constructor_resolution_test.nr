package main

import "io"

type Node = enum {
    Leaf
    Binary(left: i32, right: i32)
}

fn main() i32 {
    var n = Binary(1, 2)
    io.PrintLn("OK")
    return 0
}
