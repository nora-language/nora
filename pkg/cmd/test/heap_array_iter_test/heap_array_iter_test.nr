package main

import "io"

fn main() i32 {
    var arr = alloc i32[3]
    arr[0] = 1
    arr[1] = 2
    arr[2] = 3

    io.PrintLn("Iterating heap array:")
    for i, x in arr {
        io.PrintLn("arr[${i}] = ${x}")
    }
    return 0
}
