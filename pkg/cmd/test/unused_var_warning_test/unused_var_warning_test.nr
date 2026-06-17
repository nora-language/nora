package main

import "io"

fn test_unused(y: i32, _unused_param: i32) i32 {
    var x: i32 = 100
    var _unused_local: i32 = 200
    var z: i32 = 300
    return z
}

fn main() i32 {
    var result = test_unused(1, 2)
    io.PrintLn("result: ${result}")
    return 0
}
