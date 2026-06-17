package main

import "io"

type Node = struct {
    val: i32
}

fn main() i32 {
    var capacity = 5
    // Allocate array of pointers
    var b = alloc @Node[capacity]

    var i = 0
    while (i < capacity) {
        b[i] = none
        i = i + 1
    }

    io.PrintLn("Allocated and initialized pointer array successfully")
    return 0
}
