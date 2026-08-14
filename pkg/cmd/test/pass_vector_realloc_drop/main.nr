package main

import "collections"
import "io"

pub type Inner = struct {
    data: i32[]
}

pub fn NewInner() @Inner {
    return @Inner {
        data: alloc i32[10]
    }
}

pub fn main() {
    var vec = collections.NewVector[@Inner](2)
    
    vec.Push(NewInner())
    vec.Push(NewInner())
    
    // Attempt to initialize the first element's data
    var before = vec.GetMut(0)
    before.data[0] = 123456
    
    // Reallocate
    vec.Push(NewInner())
    
    var first = vec.GetMut(0)
    
    // Read from freed memory
    var val = first.data[0]
    io.Print("Value: ")
    io.PrintLn(val)
}
