package main

import "io"

fn print_num(val: i32) {
    io.PrintLn("Deferred loop num:", val)
}

fn test_loop_defers() {
    var i = 0
    while i < 3 {
        var num = i + 1
        defer print_num(num)
        i++
    }
}

fn main() i32 {
    io.PrintLn("Starting loop defer test")
    test_loop_defers()
    io.PrintLn("Finished loop defer test")
    return 0
}
