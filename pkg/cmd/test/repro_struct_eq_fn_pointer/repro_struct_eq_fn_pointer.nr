package main

import "io"
import "os"

pub fn add_one(x: i32) i32 {
    return x + 1
}

pub fn add_two(x: i32) i32 {
    return x + 2
}

pub type Handler = struct {
    id: i32,
    cb: fn(i32) i32
}

fn main() i32 {
    var h1 = Handler { id: 1, cb: add_one }
    var h2 = Handler { id: 1, cb: add_one }
    var h3 = Handler { id: 1, cb: add_two }

    // Structural equality check on struct containing function pointer fields!
    if !(h1 == h2) {
        io.PrintLn("Error: Struct equality with identical fn pointers failed")
        os.Exit(1)
    }

    if h1 == h3 {
        io.PrintLn("Error: Struct equality with differing fn pointers passed unexpectedly")
        os.Exit(1)
    }

    io.PrintLn("Struct eq fn pointer test passed successfully!")
    return 0
}
