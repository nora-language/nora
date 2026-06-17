package main

import "io"

extern fn malloc(size: i32) ptr

fn main() {
    io.PrintLn("Allocating via extern malloc...")
    var p = malloc(100) // This will be nr_malloc due to macro, and NEVER freed by solver because it's ptr
    io.PrintLn("Exiting without freeing...")
}
