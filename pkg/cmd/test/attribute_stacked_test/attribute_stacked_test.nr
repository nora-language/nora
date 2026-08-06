package main

import "io"
import "os"

[copyable]
pub type StackedPoint = struct {
    x: i32,
    y: i32
}

[inline]
pub fn double_val(val: i32) i32 {
    return val * 2
}

fn main() i32 {
    var p1 = StackedPoint { x: 10, y: 20 }
    var p2 = p1 // Copyable check: p1 is safely copied
    var doubled = double_val(p1.x) // Inline check: stacked attributes ([copyable] struct, [inline] function)

    if doubled != 20 || p2.y != 20 {
        io.PrintLn("Error in stacked attribute point calculation")
        os.Exit(1)
    }

    io.PrintLn("Stacked attribute test passed: doubled=", doubled)
    return 0
}
