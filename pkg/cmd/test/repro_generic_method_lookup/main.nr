package main

import "./pkg_vec"

pub type PathConstraint[T] = struct {
    id: i32
}

pub fn (c: #PathConstraint[T]) FindClosestT[T](pos: #pkg_vec.Vector3[T]) {
    var diff = pos.Sub[T](#pos)
}

pub fn main() {
    var v = pkg_vec.Vector3[f32]{ x: f32(1.0) }
    var c = PathConstraint[f32]{ id: 1 }
    c.FindClosestT[f32](#v)
}
