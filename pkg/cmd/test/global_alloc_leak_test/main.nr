package main

import "collections"

var my_vec = collections.NewVector[i32](10)

fn main() {
    my_vec.Push(42)
    my_vec.Push(100)

    if my_vec.Get(0) != 42 {
        var x = 1
    }
}
