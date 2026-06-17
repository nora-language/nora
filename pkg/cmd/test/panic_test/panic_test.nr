package main

import "io"

fn main() {
    io.PrintLn("About to panic...")
    panic("something went wrong!")
    io.PrintLn("This should never print.")
}
