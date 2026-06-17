package main

import "io"

fn divide(a: i32, b: i32) Result[i32, str] {
    if (b == 0) {
        return Err[i32, str]("division by zero")
    }
    return Ok[i32, str](a / b)
}

fn calculate_complex(a: i32, b: i32, c: i32, d: i32) Result[i32, str] {
    var x = divide(a, b)? + divide(c, d)?
    return Ok[i32, str](x)
}

fn main() i32 {
    var res1 = calculate_complex(10, 2, 20, 5)
    match res1 {
        Ok(val) => {
            io.PrintLn("Result: ${val}")
        }
        Err(msg) => {
            io.PrintLn("Error: ${msg}")
        }
        _ => {
            io.PrintLn("Unknown")
        }
    }

    var res2 = calculate_complex(10, 2, 20, 0)
    match res2 {
        Ok(val) => {
            io.PrintLn("Result: ${val}")
        }
        Err(msg) => {
            io.PrintLn("Error: ${msg}")
        }
        _ => {
            io.PrintLn("Unknown")
        }
    }

    return 0
}
