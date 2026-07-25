package main

// This test documents a known language limitation:
// Nora does NOT support structural duck-typing on unconstrained generic parameters.
// To call methods on a generic type, you MUST use an explicit protocol/interface bound.
//
// See main.nr in this same directory for the correct approach using protocol constraints.

fn Intersect[T](shape: #T) {
    var vec = shape.Support() // FAILS: type parameter 'T' has no field or method 'Support'
}

fn main() {
}
