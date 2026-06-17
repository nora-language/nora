package main

import "io"

fn main() i32 {
    var r = 0
    var c = 0
    var found_r = 0
    var found_c = 0

    while (r < 5) {
        c = 0
        while (c < 5) {
            var val = 0
            io.Scan(#val)
            if (val == 1) {
                found_r = r + 1
                found_c = c + 1
            }
            c++
        }
        r++
    }

    var dr = found_r - 3
    if (dr < 0) {
        dr = -dr
    }

    var dc = found_c - 3
    if (dc < 0) {
        dc = -dc
    }

    io.PrintLn(dr + dc)
    return 0
}
