package main

import "io"

fn main() {
    var ch = make(chan[i32])
    io.PrintLn("Waiting for value that never comes...")
    var val = <-ch // Deadlock here
    io.PrintLn("Got: ${val}")
}
