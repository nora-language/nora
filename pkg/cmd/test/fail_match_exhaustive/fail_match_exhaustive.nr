package main

import "io"

type Option = enum {
    Some(val: i32),
    None
}

fn process(opt: #Option) i32 {
    match opt {
        Some(val) => {
            return val
        }
    }
}

fn main() i32 {
    var x = Some(42)
    io.PrintLn("Some: ${process(x)}")
    return 0
}
