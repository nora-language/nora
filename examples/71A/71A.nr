package main

import "io"

extern fn strlen(s: #str) i32

fn main() i32 {
    var n = 0
    io.Scan(#n)

    while (n > 0) {
        var word = io.ScanStr()
        var len = strlen(word)
        if (len > 10) {
            // First char, length-2, last char
            io.PrintLn("${word[0]}${len - 2}${word[len - 1]}")
        } else {
            io.PrintLn(word)
        }
        n--
    }

    return 0
}
