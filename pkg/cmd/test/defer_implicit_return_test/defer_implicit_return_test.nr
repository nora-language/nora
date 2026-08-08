package main

import "io"

fn do_something() {
    defer io.PrintLn("Deferred Implicit Return!")
    io.PrintLn("Doing something...")
    // No explicit return here
}

fn main() i32 {
    do_something()
    return 0
}
