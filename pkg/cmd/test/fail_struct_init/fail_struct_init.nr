package main

import "io"

type Point = struct {
    x: i32,
    y: i32
}

fn main() i32 {
    var p = Point {
        x: 10
    }
    io.PrintLn("p.x: ${p.x}")
    return 0
}
