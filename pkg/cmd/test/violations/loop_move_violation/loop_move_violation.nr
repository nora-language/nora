package main

import "io"

type User = struct {
    name: str
}

fn main() i32 {
    var u = User {
        name: "Alice"
    }
    var i = 0
    while (i < 2) {
        io.PrintLn(u.name) // SHOULD FAIL IN TOPOLOGY SOLVER on second iteration
        var u2 = u // u is moved to u2
        i++
    }
    return 0
}
