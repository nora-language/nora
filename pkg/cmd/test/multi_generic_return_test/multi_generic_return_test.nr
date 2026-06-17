package main

import "io"

type Pair[T, U] = struct {
    first: T,
    second: U
}

fn make_pair[T, U](f: T, s: U) @Pair[T, U] {
    return alloc Pair[T, U] {
        first: f,
        second: s
    }
}

fn main() i32 {
    var p = make_pair[i32, str](42, "Universe")
    io.PrintLn("Pair: ${p.first} and ${p.second}")
    return 0
}
