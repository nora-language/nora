package main
import "./pkg_vec"
import "./pkg_aabb"

pub fn main() i32 {
    var min = pkg_vec.NewVec3[f32](f32(0.0), f32(0.0), f32(0.0))
    var max = pkg_vec.NewVec3[f32](f32(1.0), f32(1.0), f32(1.0))
    var offset = pkg_vec.NewVec3[f32](f32(0.5), f32(0.5), f32(0.5))
    var aabb = pkg_aabb.NewAABB[f32](min, max)
    var aabb2 = aabb.Translated(offset)
    return 0
}
