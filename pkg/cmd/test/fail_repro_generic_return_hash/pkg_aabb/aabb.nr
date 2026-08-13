package pkg_aabb

pub type AABB[T] = struct {
    val: T
}

pub fn NewAABB[T](v: T) AABB[T] {
    return AABB[T] { val: v }
}

pub fn (a: #AABB[T]) Dummy[T]() bool {
    return true
}
