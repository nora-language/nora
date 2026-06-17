package main

import "io"

var counter = 0

fn increment(c: &i32) {
    // Dummy
}

fn worker() {
    increment(&counter) // <-- EXPECTED SEMANTIC ERROR: cannot take mutable lease of global variable inside a concurrent context
}

fn main() i32 {
    spawn worker()
    return 0
}
