package lib

import "io"

pub type MyDrop = struct {
    val: i32
}

pub fn (self: &MyDrop) drop() {
    io.PrintLn("MyDrop dropped!")
}

pub fn CreateDrop() @MyDrop {
    return alloc MyDrop { val: 42 }
}
