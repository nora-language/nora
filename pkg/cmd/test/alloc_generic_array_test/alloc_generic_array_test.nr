package main

import "io"

type Entry[K, V] = struct {
    key: K,
    value: V
}

fn main() i32 {
    var capacity = 10
    var b = alloc @Entry[str, i32][capacity]
    io.PrintLn("Allocated array of size ${len(b)}")
    return 0
}
