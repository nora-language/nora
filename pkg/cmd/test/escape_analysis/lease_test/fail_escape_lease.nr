package main

import "io"

fn create_closure() fn() void {
    var x = 10
    var r = #x
    var f = fn() void {
        io.PrintLn("Value: ${io.FormatInt(r)}")
    }
    return f
}

fn main() i32 {
    var f = create_closure()
    f()
    return 0
}
