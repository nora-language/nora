package main

import "io"

fn main() i32 {
    // Range loops cannot have two loop variables (key, value)
    for k, v in 0..10 {
        io.PrintLn(k, v)
    }
    return 0
}
