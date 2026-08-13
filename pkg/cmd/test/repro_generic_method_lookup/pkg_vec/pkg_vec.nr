package pkg_vec

pub type Vector3[T] = struct {
    x: T
}

pub fn (v: #Vector3[T]) Sub[T](other: #Vector3[T]) @Vector3[T] {
    return alloc Vector3[T]{ x: v.x }
}
