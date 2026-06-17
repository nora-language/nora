package main

import "io"

extern fn strlen(s: #str) i32

fn main() i32 {
    var s = io.ScanStr()

    var len = strlen(s)
    var distinct = 0
    var seen = alloc byte[256]

    var i = 0
    while (i < 256) {
        seen[i] = 0
        i++
    }

    i = 0
    while (i < len) {
        var char_code = s[i]
        if (seen[char_code] == 0) {
            seen[char_code] = 1
            distinct++
        }
        i++
    }

    if (distinct % 2 == 0) {
        io.PrintLn("CHAT WITH HER!")
    } else {
        io.PrintLn("IGNORE HIM!")
    }

    return 0
}
