package main

import "io"

type Printable = interface {
    fn (self: Printable) custom_print()
}

fn print_it[T: Printable](x: T) {
    x.custom_print()
}

type Box = struct {
    val: i32
}

fn (self: Box) custom_print() {
    io.PrintLn("Box val = ${self.val}")
}

fn main() i32 {
    var b = Box {
        val: 42
    }
    print_it(b)
    return 0
}
