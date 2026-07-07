package main

import lib "./lib.nr"
import "io"

pub fn main() i32 {
    io.PrintLn("Start!")
    var d = lib.CreateDrop()
    io.PrintLn("End!")
    return 0
}
