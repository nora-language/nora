package main

import "io"

fn find_even(n: i32) Option[i32] {
    if (n % 2 == 0) {
        return Some[i32](n)
    }
    return None[i32]
}

fn double_even(n: i32) Option[i32] {
    var val = find_even(n)?
    return Some[i32](val * 2)
}

fn main() i32 {
    var res1 = double_even(4)
    match res1 {
        Some(val) => {
            io.PrintLn("Doubled: ${val}")
        }
        None => {
            io.PrintLn("Not even")
        }
        _ => {
            io.PrintLn("Unknown")
        }
    }

    var res2 = double_even(5)
    match res2 {
        Some(val) => {
            io.PrintLn("Doubled: ${val}")
        }
        None => {
            io.PrintLn("Not even")
        }
        _ => {
            io.PrintLn("Unknown")
        }
    }

    return 0
}
