package main

import "io"

pub type Speaker = interface {
    fn Speak() void
}

pub type Person = struct {
    name: str
}

pub fn (self: #Person) Speak() void {
    io.PrintLn("Hello, ${self.name}")
}

pub fn Introduce(s: Speaker) {
    s.Speak()
}

fn main() i32 {
    var p = Person { name: "test_" + "name" }
    var s: Speaker = @p
    Introduce(@s)
    return 0
}
