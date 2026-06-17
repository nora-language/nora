package main

fn main() i32 {
    var s = "hello"
    s[0] = 101 // <-- EXPECTED SEMANTIC ERROR: cannot assign to immutable string index
    return 0
}
