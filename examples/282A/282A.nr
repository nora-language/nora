package main

import "io"

extern fn strstr(haystack: str, needle: str) ptr

fn main() i32 {
    var n = 0
    io.Scan(#n)

    var x = 0

    while (n > 0) {
        var s = io.ScanStr()
        if (strstr(s, "++") != none) {
            x++
        } else {
            x--
        }
        n--
    }

    io.PrintLn(x)
    return 0
}
