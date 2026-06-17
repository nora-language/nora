package main

import "io"

fn main() i32 {
    var x = 999
    var w = fn() void {
        var y: #i32 = #x
        io.PrintLn("Value")
    }
    spawn w()
    return 0
}
