package main

import "pkg_generic"
import "pkg_aabb"
import "pkg_math"

pub fn main() {
    var aabb = pkg_aabb.NewAABB[f64](0.0)
    var dummy = aabb.Dummy[f64]()

    var b = pkg_generic.Wrapper[pkg_aabb.AABB[f64]](aabb)
}
