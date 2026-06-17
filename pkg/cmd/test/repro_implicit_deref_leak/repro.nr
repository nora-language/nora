package main

import "io"

pub type MyStruct = struct {
    val: i32
}

pub fn create_struct() @MyStruct {
    return alloc MyStruct { val: 42 }
}

pub fn take_by_value[T](_s: T) {
    io.PrintLn("Value taken by value")
}

pub fn main() {
    var _s = create_struct()
    take_by_value[MyStruct](@_s)
}
