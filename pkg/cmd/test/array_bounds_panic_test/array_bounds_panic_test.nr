package main

import "io"

fn main() {
    var arr = [10, 20, 30]

    io.PrintLn("Accessing arr[1] (should work):")
    io.PrintLn(arr[1])

    io.PrintLn("Accessing arr[3] (should PANIC):")
    io.PrintLn(arr[3])
}
