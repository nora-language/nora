package pkg_vec

pub type Vec3[T: Copy] = struct {
    x: T
    y: T
    z: T
}

pub fn NewVec3[T: Copy](x: T, y: T, z: T) Vec3[T] {
    return Vec3[T] { x: x, y: y, z: z }
}
