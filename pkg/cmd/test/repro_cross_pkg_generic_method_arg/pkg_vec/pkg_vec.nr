package pkg_vec

pub type Vec2[T: Copy] = struct {
    x: T
    y: T
}

pub fn NewVec2[T: Copy](x: T, y: T) Vec2[T] {
    return Vec2[T] { x: x, y: y }
}

// Bug: When this method is called from OUTSIDE pkg_vec with a Vec2[f32] argument,
// the compiler fails because the parameter type is "#Vec2" (local name) but the
// argument type from the calling package is "#pkg_vec_Vec2_XXXX" (monomorphized name).
pub fn (self: #Vec2[T]) Dot[T: Copy](other: #Vec2[T]) T {
    return self.x * other.x + self.y * other.y
}
