package main

import "./vec/vec"

pub fn main() i32 {
    var a = vec.NewVec2[f32](f32(10.0), f32(20.0))
    var b = vec.NewVec2[f32](f32(1.0), f32(2.0))
    var c = a + b
    var d = a + f32(5.0)
    if (c.x == f32(11.0) && d.x == f32(15.0)) {
        return 0
    }
    return 1
}
