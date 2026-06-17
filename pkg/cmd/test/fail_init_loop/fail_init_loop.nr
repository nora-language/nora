package main

import "io"

fn main() i32 {
    var x: i32
    var cond = true
    while (cond) {
        x = 42
        cond = false
    }
    io.PrintLn("x: ${x}")
    return 0
}
