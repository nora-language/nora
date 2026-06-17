package main

import "io"

type Box[T] = struct {
    value: T
}

fn main() i32 {
    var b1 = Box[i32] {
        value: 10
    }
    var b2 = Box[str] {
        value: "Hello"
    }

    io.PrintLn("b1.value = ${b1.value}")
    io.PrintLn("b2.value = ${b2.value}")

    return 0
}
