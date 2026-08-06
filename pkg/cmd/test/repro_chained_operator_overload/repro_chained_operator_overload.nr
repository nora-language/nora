package main

import "io"
import "os"

[copyable]
pub type Vec2 = struct {
    x: i32,
    y: i32
}

pub fn (self: #Vec2) operator+(other: Vec2) Vec2 {
    return Vec2 {
        x: self.x + other.x,
        y: self.y + other.y
    }
}

pub fn (self: #Vec2) operator*(other: Vec2) Vec2 {
    return Vec2 {
        x: self.x * other.x,
        y: self.y * other.y
    }
}

fn main() i32 {
    var v1 = Vec2 { x: 1, y: 2 }
    var v2 = Vec2 { x: 3, y: 4 }
    var v3 = Vec2 { x: 5, y: 6 }
    var v4 = Vec2 { x: 2, y: 2 }

    // Chained nested binary operator overloading expression
    var res = (v1 + v2) * (v3 + v4)

    // (1+3, 2+4) * (5+2, 6+2) = (4, 6) * (7, 8) = (28, 48)
    if res.x != 28 || res.y != 48 {
        io.PrintLn("Error: Chained operator overload result wrong")
        os.Exit(1)
    }

    io.PrintLn("Chained operator overload test passed: x=", res.x, "y=", res.y)
    return 0
}
