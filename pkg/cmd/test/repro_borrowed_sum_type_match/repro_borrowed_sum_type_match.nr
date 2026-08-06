package main

import "io"
import "os"

pub type IntBox = struct {
    val: i32
}

fn check_option(opt: #Option[IntBox]) i32 {
    match opt {
        Some(b) => {
            return b.val
        }
        None => {
            return -1
        }
    }
}

fn main() i32 {
    var item = Some[IntBox](IntBox { val: 99 })
    var val1 = check_option(#item)
    // Borrow check: item was read-borrowed (#), so it remains valid!
    var val2 = check_option(#item)

    if val1 != 99 || val2 != 99 {
        io.PrintLn("Error: Borrowed sum type match failed")
        os.Exit(1)
    }

    io.PrintLn("Borrowed sum type match test passed: val=", val2)
    return 0
}
