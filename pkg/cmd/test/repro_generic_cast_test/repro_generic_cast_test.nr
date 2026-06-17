package main

import "io"
import "collections"

type Position = struct {
    x: i32,
    y: i32
}

fn InitColumn[T](cap: i32) any {
    var vec = collections.NewVector[T](cap)
    var res: any = @vec
    return @res
}

fn AddRawColumn[T](col_any: &any, val: T) void {
    var vec: &collections.Vector[T] = collections.Vector[T](col_any)
    var temp: T = val
    vec.Push[T](temp)
}

fn main() i32 {
    var col_any = InitColumn[Position](10)
    var pos = Position { x: 42, y: 43 }
    AddRawColumn[Position](&col_any, pos)
    
    // check if it succeeds
    var vec_back = collections.Vector[Position](col_any)
    if (vec_back.size != 1) {
        panic("Size should be 1")
    }
    var p = vec_back.Get(0)
    if (p.x != 42) {
        panic("x should be 42")
    }
    
    io.PrintLn("Success!")
    return 0
}
