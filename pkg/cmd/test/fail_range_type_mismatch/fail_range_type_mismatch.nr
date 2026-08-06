package main

import "io"

fn main() i32 {
    var start: i32 = 0
    var end: i64 = 10
    // Range bounds start and end have mismatched types (i32 vs i64)
    for i in start..end {
        io.PrintLn("i:", i)
    }
    return 0
}
