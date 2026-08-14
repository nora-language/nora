package main

import "io"

fn main() {
    var x = 10 - 
            2 * 3 + 
            4
    
    // 10 - (2 * 3) + 4 = 10 - 6 + 4 = 8
    if x != 8 {
        io.PrintLn("FAIL: multiline arithmetic x", x)
        panic("Assertion failed")
    }

    var y = 20 /
            5 -
            1
    
    // 20 / 5 - 1 = 4 - 1 = 3
    if y != 3 {
        io.PrintLn("FAIL: multiline arithmetic y", y)
        panic("Assertion failed")
    }

    io.PrintLn("PASS")
}
