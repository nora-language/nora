package main
import "io"
export fn my_func() {}
fn main() {
    var p = ptr(my_func)
    io.PrintLn("Cast OK")
}
