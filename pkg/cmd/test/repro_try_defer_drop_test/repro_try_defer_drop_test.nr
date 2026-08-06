package main

import "io"
import "os"

fn cleanup() {
    io.PrintLn("Defer executed safely")
}

fn fail_op() Result[i32, str] {
    return Err[i32, str]("Simulated failure")
}

fn try_with_defer() Result[i32, str] {
    var vec = alloc i32[100] // Heap allocation before try
    
    defer cleanup()

    var val = fail_op()? // Will return Err early!
    return Ok[i32, str](val + 1)
}

fn main() i32 {
    match try_with_defer() {
        Ok(v) => {
            io.PrintLn("Error: Expected Err return")
            os.Exit(1)
        }
        Err(msg) => {
            io.PrintLn("Caught expected error from try+defer: ${msg}")
        }
    }
    return 0
}
