package main

import "./lib"
import "io"

extern fn test_export_function() i32

pub fn main() {
    var a = test_export_function()
    io.PrintLn("test ${a}")
}
