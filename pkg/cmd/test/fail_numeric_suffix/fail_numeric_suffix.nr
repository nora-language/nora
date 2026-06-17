package main

import "io"

fn main() i32 {
    var x = 10i64 // <-- SHOULD FAIL PARSER (type suffixes on numeric literals are disallowed)
    io.PrintLn(x)
    return 0
}
