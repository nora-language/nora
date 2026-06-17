package main

import "io"

type Option[T] = enum {
    None,
    Some(val: T)
}

fn print_opt(o: Option[i32]) {
    match o {
        Some(v) => {
            io.PrintLn("Value: ${v}")
        }
        None => {
            io.PrintLn("No value")
        }
    }
}

fn main() i32 {
    var o1 = Some[i32](10)
    var o2 = None[i32]

    print_opt(o1)
    print_opt(o2)

    return 0
}
