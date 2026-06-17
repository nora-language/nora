package main

import "collections"
import "io"

pub fn main() {
    var vecs = collections.NewVector[@collections.Vector[i32]](16)
    
    for i in 0..17 {
        var v = collections.NewVector[i32](8)
        var val1 = i
        v.Push[i32](val1)
        vecs.Push[@collections.Vector[i32]](@v)
        
        var first = #vecs.data[0]
        var unused = first.data[0]
    }
}
