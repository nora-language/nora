package main

import "pkg_vec"
import "pkg_aabb"
import "io"

pub fn main() i32 {
    var min = pkg_vec.Vector3[f32]{ x: f32(0.0), y: f32(0.0), z: f32(0.0) }
    var max = pkg_vec.Vector3[f32]{ x: f32(1.0), y: f32(1.0), z: f32(1.0) }
    
    var aabb = pkg_aabb.NewAABB[f32](min, max)
    io.PrintLn("AABB created successfully!\n")
    return 0
}
