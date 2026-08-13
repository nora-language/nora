package pkg_vec

pub type Vector3[T] = struct {
    x: T,
    y: T,
    z: T,
}

pub fn NewVector3[T](x: T, y: T, z: T) @Vector3[T] {
    return alloc Vector3[T] { x: x, y: y, z: z }
}
