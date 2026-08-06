package main

import "io"
import "os"

pub type Pair[A, B] = struct {
    first: A,
    second: B
}

pub type Container[T] = struct {
    val: T
}

// Receiver method has its own generic parameter U distinct from struct generic parameter T
pub fn (c: #Container[T]) Zip[U](other: #Container[U]) Pair[#T, #U] {
    return Pair[#T, #U] {
        first: #c.val,
        second: #other.val
    }
}

fn main() i32 {
    var c1 = Container[i32] { val: 42 }
    var c2 = Container[str] { val: "world" }

    // Method-level generic parameter U resolved independently from receiver parameter T!
    var p = c1.Zip[str](#c2)

    if p.first != 42 || p.second != "world" {
        io.PrintLn("Error: Multi generic method zip values wrong")
        os.Exit(1)
    }

    io.PrintLn("Multi generic method zip test passed: first=", p.first, "second=", p.second)
    return 0
}
