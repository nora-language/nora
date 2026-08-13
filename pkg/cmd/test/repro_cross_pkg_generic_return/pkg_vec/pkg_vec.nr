package pkg_vec

pub type Vec2[T: Copy] = struct {
    x: T
    y: T
}

pub fn NewVec2[T: Copy](x: T, y: T) Vec2[T] {
    return Vec2[T] { x: x, y: y }
}
