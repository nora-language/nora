package main

import "io"

pub type Vector = struct {
    x: i32
    y: i32
}

// 1. Operator overloading for Vector + Vector
pub fn (self: #Vector) operator+(other: Vector) Vector {
    return Vector{
        x: self.x + other.x,
        y: self.y + other.y,
    }
}

// 2. Operator overloading for Vector + Scalar (Demonstrates Implicit Overloading)
pub fn (self: #Vector) operator+(scalar: i32) Vector {
    return Vector{
        x: self.x + scalar,
        y: self.y + scalar,
    }
}

// 3. General Implicit Function Overloading
pub fn process(val: i32) {
    if val != 10 {
        panic("expected 10")
    }
    io.PrintLn("Processed i32")
}

pub fn process(val: str) {
    if val != "hello" {
        panic("expected hello")
    }
    io.PrintLn("Processed str")
}

pub fn main() {
    // Test general implicit overloading
    process(10)
    process("hello")

    // Test operator overloading (Vector + Vector)
    var v1 = Vector{x: 10, y: 20}
    var v2 = Vector{x: 5, y: 15}
    var v3 = v1 + v2

    if v3.x != 15 || v3.y != 35 {
        panic("Vector + Vector failed")
    }

    // Test operator overloading (Vector + Scalar)
    var v4 = v1 + 5
    if v4.x != 15 || v4.y != 25 {
        panic("Vector + Scalar failed")
    }

    io.PrintLn("All tests passed!")
}
