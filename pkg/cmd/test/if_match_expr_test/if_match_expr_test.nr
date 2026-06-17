package ifmatchtest

import "io"

type Status = enum {
    Ok(value: i32),
    Err(code: i32)
}

fn main() i32 {
    var x = 60

    // 1. If expression
    var coba = if (x > 50) {
        1
    } else {
        0
    }
    io.PrintLn("if-expr: ${coba}")

    // 2. If statement (no assignment, should work as before)
    if (x > 50) {
        io.PrintLn("if-stmt: worked")
    }

    // 3. Match expression
    var s = Ok(100)
    var val = match s {
        Ok(v) => {
            v
        }
        Err(c) => {
            0
        }
    }
    io.PrintLn("match-expr: ${val}")

    // 4. Match statement
    match s {
        Ok(v) => {
            io.PrintLn("match-stmt: Ok with ${v}")
        }
        Err(c) => {
            io.PrintLn("match-stmt: Err with ${c}")
        }
    }

    // 5. Nested expressions
    var nested = if (x > 50) {
        match s {
            Ok(v) => {
                v + 1
            }
            Err(c) => {
                c
            }
        }
    } else {
        -1
    }
    io.PrintLn("nested: ${nested}")

    return 0
}
