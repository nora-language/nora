package main

import "io"

type Position = struct {
    x: i32,
    y: i32
}

fn wrap_any[T](val: T) any {
    var local_val = val
    var res: any = @local_val
    return @res
}

fn main() i32 {
    io.PrintLn("--- Repro Struct Any Test Start ---")
    var pos = Position { x: 42, y: 100 }
    var pos_any = wrap_any[Position](pos)
    io.PrintLn("Wrapped successfully!")
    return 0
}
