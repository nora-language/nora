package main

import "io"

pub type DummySystem = struct {
    device: ptr,
    surface: ptr,
}

pub type World = struct { id: i32 }
pub type Transform = struct { id: i32 }

pub fn NewQuery1[T](w: &World, id: i32) ptr {
    return ptr(0)
}

pub fn DrawMeshes(w: &World, pass: ptr) {}

pub fn (self: &DummySystem) Update(w: &World) {
    var query = NewQuery1[Transform](w, 1)
    
    DrawMeshes(w, ptr(0))
}

pub fn main() {
    io.PrintLn("Test compiling...")
}
