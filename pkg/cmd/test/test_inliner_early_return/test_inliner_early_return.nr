package main

import "io"
import "os"

[inline]
fn get_val(cond: bool) i32 {
    if cond {
        return 42
    }
    return 100
}

fn main() i32 {
    var x = get_val(true)
    if x != 42 {
        io.PrintLn("Inliner bug! Early return fell through. Expected 42, got: ", x)
        os.Exit(1)
    }
    io.PrintLn("Inliner works fine.")
    return 0
}
