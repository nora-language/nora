package main

// REAL UNRESOLVED LIMITATION #1: No Structural Method Resolution (Duck-Typing) for Generics
// What you want to write:
// Call a method 'shape.Support()' directly on an unconstrained generic parameter 'T'.
//
// The Reality / Compiler Error:
// Error: type parameter 'T' has no field or method 'Support'
// Nora strictly requires explicit interface/protocol bounds on 'T'.

fn Intersect[T](shape: #T) {
    var vec = shape.Support() // FAILS: type parameter 'T' has no field or method 'Support'
}

fn main() {
}
