package main

import "io"

fn main() i32 {
    var n: i32 = 0
    io.Scan(#n)

    var total = 0
    while (n > 0) {
        var a = 0
        var b = 0
        var c = 0
        io.Scan(#a, #b, #c)
        if (a + b + c >= 2) {
            total++
        }
        n--
    }
    io.PrintLn(total)
    return 0
}
