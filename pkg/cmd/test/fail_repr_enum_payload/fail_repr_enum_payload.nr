package main

import "io"

// ERROR: [repr("i32")] cannot be applied to an enum with payload fields (CustomData)
[repr("i32")]
pub type PayloadEnum = enum {
    Empty,
    CustomData(val: i32)
}

fn main() i32 {
    var item = CustomData(42)
    io.PrintLn("This should fail semantic analysis")
    return 0
}
