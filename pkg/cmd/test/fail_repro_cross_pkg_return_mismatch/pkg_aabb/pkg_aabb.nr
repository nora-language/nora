package pkg_aabb

import "pkg_vec"

pub type AABB[T] = struct {
    min: pkg_vec.Vector3[T],
    max: pkg_vec.Vector3[T],
}

pub fn NewAABB[T](min: pkg_vec.Vector3[T], max: pkg_vec.Vector3[T]) AABB[T] {
    return AABB[T] { min: min, max: max }
}
