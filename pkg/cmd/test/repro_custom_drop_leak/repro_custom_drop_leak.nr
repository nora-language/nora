package main

import "io"

type CustomStruct = struct {
    x: i32,
    y: i32
}

pub fn (self: @CustomStruct) drop() {
    io.PrintLn("Custom drop run!")
}

fn main() i32 {
    var c = alloc CustomStruct { x: 42, y: 100 }
    return 0
}
