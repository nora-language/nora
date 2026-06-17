package main

import "io"
import "string"

fn main() {
    var s = "NoraLang"
    
    // 1. Test len()
    var length = len(s)
    if length != 8 {
        io.PrintLn("FAIL: len(s) was not 8")
    }

    // 2. Test ptr()
    var p = ptr(s)
    if p == none {
        io.PrintLn("FAIL: ptr(s) is null")
    }

    // 3. Test for-in loop
    for i, ch in s {
        io.PrintLn("${i} : ${string.FromChar(ch)}")
    }

    // 4. Test while loop with indexing []
    var i = 0
    while i < len(s) {
        var ch = s[i]
        io.PrintLn("${i} : ${string.FromChar(ch)}")
        i = i + 1
    }
}
