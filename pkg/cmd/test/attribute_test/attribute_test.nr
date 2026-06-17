package main

import "io"

[custom("arg1", "arg2")]
fn my_func() {
    io.PrintLn("hello from my_func")
}

[simple]
type MyStruct = struct {
    x: i32
}

fn main() {
    my_func()
    io.PrintLn("hello from main")
}
