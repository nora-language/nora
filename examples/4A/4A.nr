package main

import "io"

fn main() i32 {
    var w: i32 = 0
    io.Scan(#w)

    if (w > 2 && w % 2 == 0) {
        io.PrintLn("YES")
    } else {
        io.PrintLn("NO")
    }

    return 0
}
