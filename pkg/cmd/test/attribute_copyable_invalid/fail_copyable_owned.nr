package main

import "io"

// A generic vector structure that owns a memory allocation
pub type Vector[T] = struct {
    data: @T[],
    len: i32
}

// A struct containing an owned pointer
[copyable]
pub type PointData = struct {
    x: i32,
    y: i32,
    owned_vec: @Vector[i32] // ERROR: cannot mark struct as copyable if it contains an owned field
}

pub fn main() {
    io.PrintLn("This should not compile!")
}
