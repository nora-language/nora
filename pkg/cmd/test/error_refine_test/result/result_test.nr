package main

import "io"

type ErrorCode = struct {
    code: i32
}

fn validate_positive(n: i32) Result[i32, ErrorCode] {
    if (n < 0) {
        return Err[i32, ErrorCode](ErrorCode {
            code: -1
        })
    }
    return Ok[i32, ErrorCode](n)
}

fn process_numbers(a: i32, b: i32) Result[i32, ErrorCode] {
    var x = validate_positive(a)?
    var y = validate_positive(b)?
    return Ok[i32, ErrorCode](x + y)
}

fn main() i32 {
    var res1 = process_numbers(10, 20)
    match res1 {
        Ok(val) => {
            io.PrintLn("Sum: ${val}")
        }
        Err(e) => {
            io.PrintLn("Error code: ${e.code}")
        }
        _ => {
            io.PrintLn("Unknown")
        }
    }

    var res2 = process_numbers(10, -5)
    match res2 {
        Ok(val) => {
            io.PrintLn("Sum: ${val}")
        }
        Err(e) => {
            io.PrintLn("Error code: ${e.code}")
        }
        _ => {
            io.PrintLn("Unknown")
        }
    }

    return 0
}
