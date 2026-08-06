package main

import "io"

fn next_val(counter: &i32) Result[i32, str] {
    if counter == 3 {
        return Err[i32, str]("Reached maximum counter limit")
    }
    counter = counter + 1
    return Ok[i32, str](counter)
}

fn process_stream() Result[i32, str] {
    var count = 0
    var sum = 0

    // ERROR: Postfix ? operator cannot be evaluated directly inside while condition expression header
    while next_val(&count)? < 3 {
        sum = sum + count
    }

    return Ok[i32, str](sum)
}

fn main() i32 {
    match process_stream() {
        Ok(s) => io.PrintLn("s=", s),
        Err(msg) => io.PrintLn("msg=", msg)
    }
    return 0
}
