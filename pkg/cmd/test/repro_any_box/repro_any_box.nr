package main

import "io"

type Position = struct {
    x: i32,
    y: i32
}

fn receive_any(val: any) {
    // Unbox and check values
    var pos = Position(val)
    io.PrintLn("Inside receive_any:")
    io.PrintLn("pos.x = ${io.FormatInt(pos.x)}")
    io.PrintLn("pos.y = ${io.FormatInt(pos.y)}")
    
    if (pos.x != 42 || pos.y != 100) {
        panic("Value corrupted due to double-boxing!")
    }
}

fn test_any(val: any) {
    // Move the any box to a local variable
    var comp = @val
    // Pass the moved any box to a function accepting any
    receive_any(@comp)
}

fn main() i32 {
    io.PrintLn("--- Repro Any Box Test ---")
    var pos = Position { x: 42, y: 100 }
    var pos_any: any = @pos
    
    test_any(@pos_any)
    io.PrintLn("Success!")
    return 0
}
