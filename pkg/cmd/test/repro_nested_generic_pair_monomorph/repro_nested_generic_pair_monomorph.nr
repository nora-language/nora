package main

import "io"
import "os"

pub type Box[T] = struct {
    val: T
}

pub type Pair[A, B] = struct {
    first: A,
    second: B
}

fn main() i32 {
    var b1 = Box[i32] { val: 100 }
    var b2 = Box[str] { val: "nora" }

    var p = Pair[Box[i32], Box[str]] {
        first: b1,
        second: b2
    }

    if p.first.val != 100 || p.second.val != "nora" {
        io.PrintLn("Error: Nested generic pair values wrong")
        os.Exit(1)
    }

    io.PrintLn("Nested generic pair test passed: first=", p.first.val, "second=", p.second.val)
    return 0
}
