package main

import "io"

fn main() i32 {
    var x: i32
    var cond = true
    if (cond) {
        x = 42
    } else {
        x = 24
    }
    io.PrintLn("x: ${x}")
    return 0
}
