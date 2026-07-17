package main

import "io"
import "./subpkg"

pub fn Calculate(x: i32) i32 {
    return x + 5
}

fn main() i32 {
    var a = Calculate(10)
    var b = subpkg.Calculate(10)

    if (a != 15) {
        io.PrintLn("FAIL: main.Calculate returned incorrect value due to inliner collision!")
        panic("Inliner collision: main.Calculate was substituted by subpkg.Calculate")
    }

    if (b != 1000) {
        io.PrintLn("FAIL: subpkg.Calculate returned incorrect value!")
        panic("Inliner collision: subpkg.Calculate failed")
    }

    io.PrintLn("Inliner name collision test passed successfully!")
    return 0
}
