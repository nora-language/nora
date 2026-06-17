package main

import "io"

type Printable = interface {
    fn (self: #Printable) custom_print()
}

type Debuggable = interface {
    Printable
    fn (self: #Debuggable) custom_debug()
}

fn show_debug[T: Debuggable](x: #T) {
    x.custom_print()
    x.custom_debug()
}

type Box = struct {
    val: i32
}

fn (self: #Box) custom_print() {
    io.PrintLn("Box val = ${self.val}")
}

fn (self: #Box) custom_debug() {
    io.PrintLn("DEBUG: Box val = ${self.val}")
}

fn main() i32 {
    var b = Box {
        val: 123
    }

    // Box implements Debuggable because it implements custom_print() (from Printable) and custom_debug()
    show_debug(#b)

    return 0
}
