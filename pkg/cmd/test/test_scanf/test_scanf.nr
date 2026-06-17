package main

import "io"

fn main() i32 {
    var n: i32 = 0
    io.Print("Enter a number: ")
    io.Scan(#n)
    io.PrintLn("You entered: ${n}")
    return 0
}
