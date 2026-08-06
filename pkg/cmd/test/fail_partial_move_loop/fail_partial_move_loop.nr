package main

import "io"

type Wrapper = struct {
    data: @str
}

fn main() i32 {
    var w = alloc Wrapper { data: "hello" }
    
    // We loop twice
    for i in 0..2 {
        // First iteration, we move w.data.
        // Second iteration, w.data is already moved, so this should be caught by topological solver!
        var x = @w.data
        io.PrintLn("Moved:", x)
    }

    return 0
}
