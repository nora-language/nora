package main

// We can just use basic assertions or no imports for a simple pass/fail
// Alternatively, we use printf for simple tests
extern fn printf(format: str, ...) i32

pub type Vector2[T] = struct {
    x: T
    y: T
}

pub fn NewVector2[T](x: T, y: T) Vector2[T] {
    return Vector2[T] { x: x, y: y }
}

pub fn (self: #Vector2[T]) operator+[T](other: #Vector2[T]) Vector2[T] {
    return Vector2[T] { x: self.x + other.x, y: self.y + other.y }
}

pub fn (self: #Vector2[T]) operator-[T](other: #Vector2[T]) Vector2[T] {
    return Vector2[T] { x: self.x - other.x, y: self.y - other.y }
}

pub fn main() i32 {
    var a = NewVector2[f32](f32(1.0), f32(1.0))
    var b = NewVector2[f32](f32(2.0), f32(2.0))
    var c = a + b
    var d = a - b
    
    if (c.x != f32(3.0) || c.y != f32(3.0)) {
        return 1
    }
    
    if (d.x != f32(-1.0) || d.y != f32(-1.0)) {
        return 1
    }
    
    return 0
}
