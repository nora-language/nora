package main

import "io"

extern fn strlen(s: #str) i32

fn main() i32 {
    var s = io.ScanStr()

    var len = strlen(s)
    var count1 = 0
    var count2 = 0
    var count3 = 0

    var i = 0
    while (i < len) {
        var c = s[i]
        if (c == 1) {
            count1++
        }
        if (c == 2) {
            count2++
        }
        if (c == 3) {
            count3++
        }
        i++
    }

    var first = true
    while (count1 > 0) {
        if (!first) {
            io.Print("+")
        }
        io.Print("1")
        first = false
        count1--
    }
    while (count2 > 0) {
        if (!first) {
            io.Print("+")
        }
        io.Print("2")
        first = false
        count2--
    }
    while (count3 > 0) {
        if (!first) {
            io.Print("+")
        }
        io.Print("3")
        first = false
        count3--
    }
    io.PrintLn()

    return 0
}
