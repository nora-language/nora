package main

import "io"

type Container = struct {
    payload: @str
}

fn main() i32 {
    var c = alloc Container { payload: "initial" }
    
    // Partially move payload
    var p1 = @c.payload
    io.PrintLn("Moved payload:", p1)

    // Re-assign payload to restore struct validity
    c.payload = "restored"
    io.PrintLn("Restored payload:", c.payload)

    return 0
}
