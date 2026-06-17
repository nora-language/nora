package main

import "io"

type Resource = struct {
    id: i32
}

fn (self: #Resource) drop() {
    io.PrintLn("Dropping resource ${self.id}")
}

fn do_async_c_stuff(r: #Resource) {
    // Simulating async stuff
    io.PrintLn("C async call using pointer...")
}

fn main() i32 {
    io.PrintLn("Creating resource")
    var r = Resource {
        id: 42
    }

    io.PrintLn("Pinning resource")
    pin r
    do_async_c_stuff(r)
    io.PrintLn("After async call, before scope ends")

    return 0
}
