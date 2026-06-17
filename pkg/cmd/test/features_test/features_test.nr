package main

import "io"

fn main() {
    var fruits = ["apple", "banana", "cherry"]

    io.PrintLn("Fruits list:")
    for f in fruits {
        io.PrintLn(f)
    }

    io.PrintLn("Fruits with index:")
    for i, f in fruits {
        io.PrintLn(f)
    }
}
