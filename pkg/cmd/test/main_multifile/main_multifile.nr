package main

import "io"
import "pkg/cmd/test/math"

fn main() {
    var a = 10
    var b = 20
    var sum = math.add(a, b)
    var diff = math.sub(a, b)

    io.PrintLn("Sum: ${sum}")
    io.PrintLn("Diff: ${diff}")
}
