package main

import "io"
import "string"

fn main() {
    var greeting = "Hello, Nora!"

    io.PrintLn("Message:", greeting)
    io.PrintLn("Length:", string.Length(greeting))

    var is_hello = string.Equals(greeting, "Hello")
    if (is_hello) {
        io.PrintLn("It is exactly 'Hello'")
    } else {
        io.PrintLn("It is not exactly 'Hello'")
    }
}
