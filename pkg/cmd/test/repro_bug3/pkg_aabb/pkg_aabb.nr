package pkg_aabb
import "../pkg_vec"

pub type AABB[T: Copy] = struct {
    min: pkg_vec.Vec3[T]
    max: pkg_vec.Vec3[T]
}

pub fn NewAABB[T: Copy](min: pkg_vec.Vec3[T], max: pkg_vec.Vec3[T]) AABB[T] {
    return AABB[T] { min: min, max: max }
}

pub fn (aabb: #AABB[T]) Translated(offset: pkg_vec.Vec3[T]) AABB[T] {
    return AABB[T] { min: offset, max: offset }
}
