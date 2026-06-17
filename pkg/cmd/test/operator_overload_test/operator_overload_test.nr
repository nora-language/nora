package main

import "cmp"

pub type Vector2D = struct {
    x: i32
    y: i32
}

pub fn (self: #Vector2D) add(other: Vector2D) Vector2D {
    return Vector2D{
        x: self.x + other.x,
        y: self.y + other.y,
    }
}

pub fn (self: #Vector2D) sub(other: Vector2D) Vector2D {
    return Vector2D{
        x: self.x - other.x,
        y: self.y - other.y,
    }
}

pub fn (self: #Vector2D) eq(other: #Vector2D) bool {
    return self.x == other.x && self.y == other.y
}

pub fn main() {
    var v1 = Vector2D{x: 10, y: 20}
    var v2 = Vector2D{x: 5, y: 15}

    var v3 = v1 + v2
    if v3.x != 15 || v3.y != 35 {
        panic("addition failed")
    }

    var v4 = v1 - v2
    if v4.x != 5 || v4.y != 5 {
        panic("subtraction failed")
    }

    var v5 = Vector2D{x: 10, y: 20}
    if !(v1 == v5) {
        panic("equality failed")
    }

    if v1 == v2 {
        panic("inequality failed")
    }
}
