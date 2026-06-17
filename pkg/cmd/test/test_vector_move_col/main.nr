package main

import "collections"
import "io"

type Position = struct {
    x: i32,
    y: i32
}

pub fn MoveColumn[T](src_any: &any, src_row: i32, dest_any: &any) {
    var src_vec: &collections.Vector[T] = collections.Vector[T](src_any)
    var dest_vec: &collections.Vector[T] = collections.Vector[T](dest_any)
    
    var val: T = @src_vec.data[src_row]
    dest_vec.Push[T](val)
}

pub fn main() {
    var vec1 = collections.NewVector[Position](2)
    vec1.Push(Position { x: 10, y: 20 })
    vec1.Push(Position { x: 100, y: 200 })
    
    var vec2 = collections.NewVector[Position](2)
    
    var any1: any = @vec1
    var any2: any = @vec2
    
    MoveColumn[Position](&any1, 1, &any2)
    MoveColumn[Position](&any1, 0, &any2)
    
    var dest_vec: &collections.Vector[Position] = collections.Vector[Position](&any2)
    
    if (dest_vec.size != 2) {
        panic("Size mismatch")
    }
    
    if (dest_vec.data[0].x != 100 || dest_vec.data[0].y != 200) {
        panic("Dest element 0 mismatch")
    }
    
    if (dest_vec.data[1].x != 10 || dest_vec.data[1].y != 20) {
        panic("Dest element 1 mismatch")
    }
    
    io.PrintLn("MoveColumn tested successfully!")
}
