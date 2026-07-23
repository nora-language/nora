package main

import "io"

// This struct is constrained with [T: Copy] which informs the semantic analyzer
// that T is not a strict linear type.
pub type SafeVector[T: Copy] = struct {
    x: T,
    y: T,
    z: T,
}

pub fn NewSafeVector[T: Copy](x: T, y: T, z: T) @SafeVector[T] {
    return alloc SafeVector[T] { x: x, y: y, z: z }
}

pub fn (v: #SafeVector[T]) Add[T: Copy](other: #SafeVector[T]) @SafeVector[T] {
    // If T were not constrained to Copy, reusing x, y, and z would throw a move error.
    return alloc SafeVector[T] {
        x: v.x + other.x,
        y: v.y + other.y,
        z: v.z + other.z,
    }
}

pub fn main() {
    var zero = 0.0
    // Because T is constrained to Copy, and f64 is a copyable primitive, 
    // we can pass 'zero' multiple times without 'use of moved value' errors.
    // There is no need for the (zero - zero) hack here!
    var v1 = NewSafeVector[f64](zero, zero, zero)
    
    // We can also initialize using a struct literal directly:
    var v2 = alloc SafeVector[f64] { x: zero, y: zero, z: zero }
    
    var v3 = v1.Add[f64](#v2)
    
    io.PrintLn("Successfully compiled and executed with [T: Copy]!")
}
