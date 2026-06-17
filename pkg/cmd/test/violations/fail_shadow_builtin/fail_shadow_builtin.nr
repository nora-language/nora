package main

fn main() i32 {
    var i32 = 42 // <-- EXPECTED SEMANTIC ERROR: cannot shadow built-in type 'i32' with variable definition
    return 0
}
