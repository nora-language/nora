package main

import "io"
import "os"

pub type Box[T] = struct {
    value: T
}

fn main() i32 {
    // 1. Primitive instantiation: Box[Option[i32]]
    var opt_prim: Option[i32] = Some[i32](42)
    var box_prim = Box[Option[i32]] { value: opt_prim }

    match box_prim.value {
        Some(val) => {
            if val != 42 {
                io.PrintLn("Error: Primitive nested generic sum value mismatch")
                os.Exit(1)
            }
        }
        None => {
            io.PrintLn("Error: Expected Some")
            os.Exit(1)
        }
    }

    // 2. String instantiation: Box[Option[str]]
    var opt_str: Option[str] = Some[str]("hello nora")
    var box_str = Box[Option[str]] { value: opt_str }

    match box_str.value {
        Some(msg) => {
            io.PrintLn("Nested generic sum monomorphization passed: ${msg}")
        }
        None => {
            io.PrintLn("Error: Expected Some")
            os.Exit(1)
        }
    }

    return 0
}
