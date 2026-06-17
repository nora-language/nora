package main

type Position = struct {
    x: i32,
    y: i32
}

pub type Command = struct {
    comp_any: any
}

fn test_func(c: Command) {
    var a_ref: &any = &c.comp_any
    var pos_ref: &Position = Position(a_ref)
    
    if (pos_ref.x != 10) {
        panic("x mismatch!")
    }
    if (pos_ref.y != 20) {
        panic("y mismatch!")
    }
}

fn AddDeferred[T](comp: T) {
    var val = comp
    var val_any: any = val
    var cmd = Command { comp_any: @val_any }
    test_func(cmd)
}

pub fn main() {
    var p = Position { x: 10, y: 20 }
    AddDeferred[Position](p)
}
