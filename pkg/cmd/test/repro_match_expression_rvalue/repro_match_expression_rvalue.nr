package main

import "io"
import "os"

pub type DataBox = struct {
    value: i32
}

pub type Container = enum {
    ItemA(val: DataBox),
    ItemB(val: DataBox)
}

fn main() i32 {
    var c = ItemA(DataBox { value: 42 })

    // Using match directly as an R-value expression
    var extracted = match c {
        ItemA(box) => box,
        ItemB(box) => box
    }

    if extracted.value != 42 {
        io.PrintLn("Error: R-value match value wrong")
        os.Exit(1)
    }

    io.PrintLn("Match expression rvalue test passed: ${extracted.value}")
    return 0
}
