package main

// REAL UNRESOLVED LIMITATION #4: No Parameter Names in Function Type Aliases
// What you want to write:
// Documenting function type aliases with parameter names: fn(level: i32, message: str)
//
// The Reality / Compiler Error:
// Error: function types cannot have parameter names
// Function signature aliases can only declare parameter types: fn(i32, str).

pub type Callback = fn(val: i32, msg: str) // FAILS

pub fn main() {
}
