package main

import "std/collections"
import "io"

pub type Dummy = struct {
    id: i32,
}

pub fn main() {
    io.PrintLn("Start")
    var vec = collections.NewVector[Dummy](4)
    vec.Push(Dummy{id: 1})
    vec.Push(Dummy{id: 2})
    vec.Push(Dummy{id: 3})
    vec.Push(Dummy{id: 4})
    
    vec.ParMap[Dummy](fn(val: &Dummy) {
        if (val.id == 2 || val.id == 3) {
            return
        }
        io.PrintLn("Processed a value")
    })
    
    io.PrintLn("End")
}
