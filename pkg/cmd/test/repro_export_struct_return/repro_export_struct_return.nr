package main

import "io"
import "os"

pub type ExportPoint = struct {
    x: i32,
    y: i32
}

pub export fn make_export_point(x: i32, y: i32) ExportPoint {
    return ExportPoint { x: x, y: y }
}

fn main() i32 {
    var p = make_export_point(10, 20)
    if p.x != 10 || p.y != 20 {
        io.PrintLn("Error: Exported function returning struct by value failed")
        os.Exit(1)
    }

    io.PrintLn("Export struct return test passed: x=", p.x, "y=", p.y)
    return 0
}
