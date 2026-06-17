package main

import "io"

extern fn strlen(s: #str) i32

extern fn tolower(c: i32) i32

fn main() i32 {
    var s1 = io.ScanStr()
    var s2 = io.ScanStr()

    var len = strlen(s1)
    var i = 0
    var res = 0

    while (i < len) {
        var c1 = tolower(s1[i])
        var c2 = tolower(s2[i])

        if (c1 < c2) {
            res = -1
            break;
        }
        if (c1 > c2) {
            res = 1
            break;
        }
        i++
    }

    io.PrintLn(res)
    return 0
}
