package main

import "io"
import "string"

type unusedStruct = struct {
    x: i32
}

type UsedStruct = struct {
    y: i32
}

fn unused_helper() {
    io.PrintLn("Unused")
}

fn used_helper() i32 {
    var u = UsedStruct {
        y: 42
    }
    return u.y
}

fn main() i32 {
    io.PrintLn("Result: ${used_helper()}")
    return 0
}
