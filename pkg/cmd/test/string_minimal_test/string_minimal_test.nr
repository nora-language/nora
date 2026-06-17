package main

import "io"

fn test_inline_if(x: i32) bool {
    if (x > 0) {
        return false
    }
    return true
}

fn main() i32 {
    io.PrintLn("ok")
    return 0
}
