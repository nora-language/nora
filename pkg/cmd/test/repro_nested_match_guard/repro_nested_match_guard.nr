package main

import "io"
import "os"

pub type Status = enum {
    Active(code: i32),
    Pending(code: i32),
    Closed
}

fn process_status(s: Status) i32 {
    match s {
        Active(code) => {
            return match code {
                10 => 100,
                20 => 200,
                _  => 999
            }
        }
        Pending(code) => {
            return match code {
                5 => 50,
                _ => 500
            }
        }
        _ => {
            return 0
        }
    }
}

fn main() i32 {
    var s1 = process_status(Active(10))
    var s2 = process_status(Active(99))
    var s3 = process_status(Pending(5))
    var s4 = process_status(Closed)

    if s1 != 100 || s2 != 999 || s3 != 50 || s4 != 0 {
        io.PrintLn("Error: Nested match guard result wrong")
        os.Exit(1)
    }

    io.PrintLn("Nested match guard test passed: s1=", s1, "s2=", s2, "s3=", s3, "s4=", s4)
    return 0
}
