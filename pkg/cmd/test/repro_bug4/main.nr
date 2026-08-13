package main

import "./pkg_vec"

pub fn Hash2DVec(p: #pkg_vec.Vec2[f32]) pkg_vec.Vec2[f32] {
    return pkg_vec.NewVec2[f32](f32(2.0), f32(2.0))
}

pub fn main() {
    var v00 = pkg_vec.NewVec2[f32](f32(1.0), f32(1.0))
    var p00 = pkg_vec.NewVec2[f32](f32(0.0), f32(0.0))
    var g00 = Hash2DVec(#p00)
    var n00 = g00.Dot(#v00)
}
