package main

import "io"

type CustomOption = enum {
    CustomSome(val: i32),
    CustomNone
}

fn process(opt: #CustomOption) i32 {
    match opt {
        CustomSome(val) => {
            return val
        }
        CustomNone => {
            return 0
        }
    }
}

fn main() i32 {
    var x = CustomSome(42)
    var y = CustomNone

    io.PrintLn("Some: ${process(x)}")
    io.PrintLn("None: ${process(y)}")
    return 0
}
