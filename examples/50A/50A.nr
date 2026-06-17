package main

import "io"

fn main() i32 {
    var m = 0
    var n = 0
    io.Scan(#m, #n)
    io.PrintLn((m * n) / 2)
    return 0
}
