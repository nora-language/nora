package main

import "io"
import "os"

[copyable]
pub type Point[T: Copy] = struct {
    x: T,
    y: T
}

pub fn (self: Point[T]) get_x[T: Copy]() T {
    return self.x
}

pub fn (self: Point[T]) get_y[T: Copy]() T {
    return self.y
}

fn main() i32 {
    var p_i32 = Point[i32] { x: 10, y: 20 }
    var x1 = p_i32.get_x[i32]()
    var y1 = p_i32.get_y[i32]()

    var p_f64 = Point[f64] { x: 1.5, y: 2.5 }
    var x2 = p_f64.get_x[f64]()

    if x1 != 10 || y1 != 20 || x2 != 1.5 {
        io.PrintLn("Error: Generic copyable receiver method failed")
        os.Exit(1)
    }

    io.PrintLn("Generic copyable receiver test passed: x1=", x1, "x2=", x2)
    return 0
}
