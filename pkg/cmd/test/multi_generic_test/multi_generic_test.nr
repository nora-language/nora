package main

import "io"

type Pair[T, U] = struct {
    first: T,
    second: U
}

fn main() i32 {
    var p = Pair[i32, str] {
        first: 42,
        second: "Universe"
    }

    io.PrintLn("Pair: ${p.first} and ${p.second}")

    return 0
}
