package main

import "io"

type Printable = interface {
    fn (self: Printable) custom_print()
}

fn print_it[T: Printable](x: T) {
    x.custom_print()
}

type PlainStruct = struct {
    val: i32
}

fn main() i32 {
    var p = PlainStruct {
        val: 100
    }
    // ERROR: PlainStruct does not implement Printable interface
    print_it(p)
    return 0
}
