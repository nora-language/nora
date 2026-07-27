package vec

pub type Vec2[T: Copy] = struct {
    x: T,
    y: T,
}

pub fn NewVec2[T: Copy](x: T, y: T) Vec2[T] {
    return Vec2[T] { x: x, y: y }
}

pub fn (self: #Vec2[T]) operator+(other: Vec2[T]) Vec2[T] {
    return Vec2[T] { x: self.x + other.x, y: self.y + other.y }
}

pub fn (self: #Vec2[T]) operator+(scalar: T) Vec2[T] {
    return Vec2[T] { x: self.x + scalar, y: self.y + scalar }
}
