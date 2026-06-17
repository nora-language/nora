package main

import "io"

extern fn malloc(size: i32) str

fn main() i32 {
    io.PrintLn("=== Testing scan and scanln built-ins ===")

    // scan a single integer
    var n: i32 = 0
    io.PrintLn("Enter an integer:")
    io.Scan(#n)
    io.PrintLn("Read integer:", n)

    // scan two integers
    var a: i32 = 0
    var b: i32 = 0
    io.PrintLn("Enter two integers:")
    io.Scan(#a, #b)
    io.PrintLn("Read integers:", a, b)

    // scan a float
    var f: f64 = 0.0
    io.PrintLn("Enter a float:")
    io.Scan(#f)
    io.PrintLn("Read float:", f)

    // scan a string into a pre-allocated buffer
    var buf = malloc(256)
    io.PrintLn("Enter a string:")
    io.Scan(buf)
    io.PrintLn("Read string:", buf)

    io.PrintLn("All scan tests passed.")

    return 0
}
