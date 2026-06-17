package main

import "io"

var counter = 0

fn worker() {
    counter = counter + 1 // <-- EXPECTED SEMANTIC ERROR: cannot mutate global variable inside a concurrent context
}

fn main() i32 {
    spawn worker()
    return 0
}
