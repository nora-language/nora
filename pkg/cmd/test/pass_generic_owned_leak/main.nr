package main

import "collections"

pub type Sys = struct { x: i32 }

pub fn main() {
    var v1 = collections.NewVector[str](4)
    v1.Push[str]("hello")
    v1.Push[str]("world")

    var v2 = collections.NewVector[@Sys](4)
    v2.Push[@Sys](alloc Sys{x: 10})
    v2.Push[@Sys](alloc Sys{x: 20})
}
