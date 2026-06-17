package main
import "io"

/// coba from docs
[include("mylib.h")]
extern fn hello_from_c(msg: str)

fn main() {
    io.println("Nora says: Hello")
    hello_from_c("Hello from FFI!")
}
