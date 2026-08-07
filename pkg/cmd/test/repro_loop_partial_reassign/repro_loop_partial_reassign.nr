package main

import "io"

type Resource = struct {
    first: str,
    second: str
}

fn main() i32 {
    var r = Resource { first: "hello", second: "world" }
    var i = 0
    while (i < 2) {
        var f = @r.first
        io.PrintLn("Loop iteration:", i, "Moved first:", f)
        
        // Re-assigning r.first restores full struct liveness for next iteration
        r.first = "new_first"
        i = i + 1
    }

    io.PrintLn("Final first:", r.first)
    return 0
}
