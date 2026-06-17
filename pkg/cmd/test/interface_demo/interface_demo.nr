package main

import "io"

type Greeter = interface {
    fn say_hello()
}

type User = struct {
    name: str
}

fn (self: #User) say_hello() {
    io.PrintLn("Hello, my name is ${self.name}")
}

fn main() {
    var u = User {
        name: "Alice"
    }

    // Structural Typing check: assign User to Greeter
    var g: Greeter = u

    io.PrintLn("Interface assignment check passed!")
}
