package main

import "io"
import "collections"

type Position = struct {
    x: i32,
    y: i32
}

type Runner = interface {
    fn Run(cb: &Position) void
}

type MyRunner = struct {}

fn (r: #MyRunner) Run(cb: &Position) void {
    io.Print("Inside Run: x=")
    io.Print(io.FormatInt(cb.x))
    io.PrintLn("")
    cb.x = 99
}

fn main() i32 {
    io.PrintLn("--- Repro Test Start ---")

    // 1. Replicate interface pointer indirection bug
    var pos = Position { x: 10, y: 20 }
    var runner: Runner = MyRunner {}
    
    // We pass &pos (mutable lease of owned Position) to interface method Update
    io.PrintLn("Calling interface method...")
    runner.Run(&pos)
    io.Print("After interface method: x=")
    io.Print(io.FormatInt(pos.x))
    io.PrintLn("")
    if (pos.x != 99) {
        panic("Interface argument pointer dereference failed!")
    }

    // 2. Replicate any wrapping array-decay bug
    io.PrintLn("Testing any-wrapping with Vector...")
    var vec = collections.NewVector[Position](10)
    vec.Push[Position](Position { x: 1, y: 2 })

    var vec_any: any = @vec
    
    // Cast it back
    var vec_back = collections.Vector[Position](vec_any)
    io.Print("Vector back size: ")
    io.Print(io.FormatInt(vec_back.size))
    io.PrintLn("")
    
    var pos_zero = vec_back.Get(0)
    io.Print("Position 0: x=")
    io.Print(io.FormatInt(pos_zero.x))
    io.PrintLn("")
    if (pos_zero.x != 1) {
        panic("Vector any casting decayed to stack temp address!")
    }

    io.PrintLn("--- Repro Test Passed ---")
    return 0
}
