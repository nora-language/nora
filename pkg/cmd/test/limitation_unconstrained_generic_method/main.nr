package main

import "io"

// Nora requires explicit protocol/interface bounds on generic type parameters.
// Duck-typing (calling methods on an unconstrained 'T') is not supported —
// this is by design, consistent with Rust traits, Go interfaces, and Swift protocols.
//
// SOLUTION: Define a named interface and constrain 'T' to it using [T: InterfaceName].

pub type Supportable = interface {
    fn Support() i32
}

fn Intersect[T: Supportable](shape: #T) i32 {
    return shape.Support()
}

pub type Circle = struct {
    radius: i32
}

pub fn (c: #Circle) Support() i32 {
    return c.radius
}

pub fn main() {
    var c = Circle { radius: 42 }
    var result = Intersect[Circle](#c)
    io.PrintLn("Support result: ${result}")
}
