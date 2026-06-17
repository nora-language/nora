package main

import "io"

type Greeter = interface {
    fn greet() void
}

type User = struct {
    name: str
}

fn (self: #User) greet() {
    io.PrintLn("Hello, I am ${self.name}")
}

fn main() i32 {
    var u = User {
        name: "Alice"
    }

    // Test: Can we assign a concrete struct to an interface variable?
    var g: Greeter = u
    g.greet()

    return 0
}
