package main

import "collections"
import "io"

type Position = struct {
    x: i32,
    y: i32
}

fn SetRawColumn[T](col_any: &any, row: i32, val_any: &any) {
    var vec: &collections.Vector[T] = collections.Vector[T](col_any)
    var val: &T = T(val_any)
    var temp: T = @val
    vec.data[row] = temp
}

pub fn main() {
    var vec = collections.NewVector[Position](1)
    vec.Push(Position { x: 10, y: 20 })
    
    // Force resize
    vec.Push(Position { x: 100, y: 200 })
    vec.Push(Position { x: 1000, y: 2000 })
    
    if (vec.size != 3) {
        panic("Size mismatch")
    }
    
    if (vec.data[0].x != 10 || vec.data[0].y != 20) {
        panic("Element 0 mismatch")
    }
    
    if (vec.data[1].x != 100 || vec.data[1].y != 200) {
        panic("Element 1 mismatch")
    }
    
    if (vec.data[2].x != 1000 || vec.data[2].y != 2000) {
        panic("Element 2 mismatch")
    }
    
    // Now test through the any wrapper and array bounds check
    var vec_any: any = @vec
    var p = Position { x: 9999, y: 8888 }
    var p_any: any = p
    
    SetRawColumn[Position](&vec_any, 1, &p_any)
    
    var vec_ref: &collections.Vector[Position] = collections.Vector[Position](&vec_any)
    if (vec_ref.data[1].x != 9999 || vec_ref.data[1].y != 8888) {
        panic("SetRawColumn failed to update array correctly!")
    }
    
    io.PrintLn("Vector resize and array header access passed successfully.")
}
