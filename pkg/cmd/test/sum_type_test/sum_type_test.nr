package main

import "io"

type Status = enum {
    Inactive,
    Active(val: i32),
    Pending(msg: str)
}

fn describe(s: Status) {
    match s {
        Inactive => {
            io.PrintLn("Status is Inactive")
        }
        Active(val) => {
            io.PrintLn("Status is Active with value: ${val}")
        }
        Pending(msg) => {
            io.PrintLn("Status is Pending with message: ${msg}")
        }
        _ => {
            io.PrintLn("Unknown status")
        }
    }
}

fn main() i32 {
    var s1 = Inactive
    var s2 = Active(42)
    var s3 = Pending("Wait for it")

    describe(s1)
    describe(s2)
    describe(s3)

    return 0
}
