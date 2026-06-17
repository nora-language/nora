package main

import "io"

extern fn strlen(s: str) i32

fn main() i32 {
    var n = 0
    io.Scan(#n)
    while (n > 0) {
        var word = io.ScanStr()
        var l = strlen(word)
        io.PrintLn(l)
        n--
    }
    return 0
}
