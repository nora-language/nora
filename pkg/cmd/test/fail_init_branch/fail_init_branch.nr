package main

import "io"

fn main() i32 {
    var x: i32
    var cond = true
    if (cond) {
        x = 42
    }
    io.PrintLn("x: ${x}")
    return 0
}
