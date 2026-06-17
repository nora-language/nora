package main

import "io"

fn test_defer() {
    io.PrintLn("Start test_defer")
    defer io.PrintLn("Deferred 1")
    defer io.PrintLn("Deferred 2")
    io.PrintLn("End test_defer")
}

fn test_defer_return() i32 {
    defer io.PrintLn("Cleanup before return")
    io.PrintLn("Returning 42")
    return 42
}

fn main() i32 {
    test_defer()
    var x = test_defer_return()
    io.PrintLn("Got ${x}")
    return 0
}
