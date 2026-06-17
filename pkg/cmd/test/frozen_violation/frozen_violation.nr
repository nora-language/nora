package main

import "io"

fn main() i32 {
    var x = 10
    var f = fn() void {
        io.PrintLn("Closure: x = ${io.FormatInt(x)}")
    }

    // This should FAIL because x is captured by closure f (and thus frozen)
    x = 20

    f()
    return 0
}
