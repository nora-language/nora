package main

import "io"
import "os"

pub fn apply_op[T, U](f: fn(T) U, val: T) U {
    return f(val)
}

pub fn increment(x: i32) i32 {
    return x + 1
}

pub fn stringify(x: i32) str {
    return "val_${x}"
}

fn main() i32 {
    var num_res = apply_op[i32, i32](increment, 10)
    var str_res = apply_op[i32, str](stringify, 42)

    if num_res != 11 || str_res != "val_42" {
        io.PrintLn("Error: Higher order generic function result wrong")
        os.Exit(1)
    }

    io.PrintLn("Higher-order generic function test passed: num=", num_res, "str=", str_res)
    return 0
}
