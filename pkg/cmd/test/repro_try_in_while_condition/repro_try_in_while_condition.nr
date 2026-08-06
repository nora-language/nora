package main

import "io"
import "os"

fn next_val(counter: &i32) Result[i32, str] {
    if *counter == 3 {
        return Err[i32, str]("Reached maximum counter limit")
    }
    *counter = *counter + 1
    return Ok[i32, str](*counter)
}

fn process_stream() Result[i32, str] {
    var count = 0
    var sum = 0

    // Postfix ? operator used directly inside while loop condition expression
    while next_val(&count)? < 10 {
        sum = sum + count
    }

    return Ok[i32, str](sum)
}

fn main() i32 {
    match process_stream() {
        Ok(s) => {
            io.PrintLn("Error: Expected Err early return from try in while condition")
            os.Exit(1)
        }
        Err(msg) => {
            io.PrintLn("Caught expected error from try in while condition: ${msg}")
        }
    }
    return 0
}
