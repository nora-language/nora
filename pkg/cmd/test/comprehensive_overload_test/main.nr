package main

import "std/io"

// --- 1. Function Overloading ---
pub fn multiply(a: i32, b: i32) i32 {
    return a * b
}

pub fn multiply(a: f64, b: f64) f64 {
    return a * b
}

pub fn multiply(a: str, b: i32) str {
    // Just a dummy return to test routing
    pin b
    return a
}


// --- 2. Method Overloading ---
pub type Logger = struct {
    count: i32
}

pub fn (l: &Logger) log(val: i32) {
    pin val
    l.count = l.count + 1
}

pub fn (l: &Logger) log(val: str) {
    pin val
    l.count = l.count + 1
}

pub fn (l: &Logger) log(val: f64) {
    pin val
    l.count = l.count + 1
}


// --- 3. Operator Overloading (Standard Struct) ---
pub type Vec2 = struct {
    x: i32
    y: i32
}

pub fn (self: #Vec2) operator+(other: Vec2) Vec2 {
    return Vec2 {
        x: self.x + other.x,
        y: self.y + other.y
    }
}

pub fn (self: #Vec2) operator-(other: Vec2) Vec2 {
    return Vec2 {
        x: self.x - other.x,
        y: self.y - other.y
    }
}

// Overloading the * operator for scalar multiplication
pub fn (self: #Vec2) operator*(scalar: i32) Vec2 {
    return Vec2 {
        x: self.x * scalar,
        y: self.y * scalar
    }
}

pub fn (self: #Vec2) operator==(other: #Vec2) bool {
    return self.x == other.x && self.y == other.y
}

pub fn (self: #Vec2) operator!=(other: #Vec2) bool {
    return self.x != other.x || self.y != other.y
}


// --- 4. Operator Overloading (Generic Struct) ---
pub type Box[T] = struct {
    val: T
}

pub fn (self: #Box[T]) operator+(other: Box[T]) Box[T] {
    // We'll just assume T is i32 for the sake of the test, 
    // actually Nora generics might complain if T is not constrained and we use +.
    // But since Box is generic, let's just return a new Box.
    // Wait, let's use a known type in the method or just return other.
    return Box[T] { val: other.val }
}

pub fn (self: #Box[T]) operator==(other: #Box[T]) bool {
    // Generic equality might need Eq constraint, but let's just return true for test
    return true
}


// --- Main Execution ---
pub fn main() {
    // Test Function Overload
    var res1 = multiply(10, 5)
    if res1 != 50 { panic("multiply i32 failed") }

    var res2 = multiply(10.5, 2.0)
    if res2 < 20.0 || res2 > 22.0 { panic("multiply f64 failed") }
    
    var res3 = multiply("hello", 5)
    pin res3

    // Test Method Overload
    var logger = Logger { count: 0 }
    logger.log(42)
    logger.log("test")
    logger.log(3.14)
    if logger.count != 3 { panic("method overload failed") }

    // Test Operator Overload (Standard)
    var v1 = Vec2 { x: 10, y: 20 }
    var v2 = Vec2 { x: 5, y: 5 }
    
    var v3 = v1 + v2
    if v3.x != 15 || v3.y != 25 { panic("operator+ failed") }
    
    var v4 = v1 - v2
    if v4.x != 5 || v4.y != 15 { panic("operator- failed") }
    
    var v5 = v1 * 2
    if v5.x != 20 || v5.y != 40 { panic("operator* failed") }
    
    if v1 == v2 { panic("operator== failed") }
    if v1 != v1 { panic("operator!= failed") }

    // Test Operator Overload (Generic)
    var b1 = Box[i32] { val: 10 }
    var b2 = Box[i32] { val: 20 }
    
    var b3 = b1 + b2
    if b3.val != 20 { panic("generic operator+ failed") }
    
    if !(b1 == b2) { panic("generic operator== failed") }
}
