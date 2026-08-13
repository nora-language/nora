package pkg_rect

import "../pkg_vec"

// Rect contains Vec2 fields — a nested generic struct.
pub type Rect[T: Copy] = struct {
    min: pkg_vec.Vec2[T]
    max: pkg_vec.Vec2[T]
}

// Bug: When NewRect is called from outside pkg_rect (e.g. from main),
// the compiler monomorphizes the return type as `pkg_rect_Rect_XXXXXXXX`,
// but the body constructs `Rect[T]` (local name).
// Error: "cannot return value of type @Rect from function returning @pkg_rect_Rect_XXXXXXXX"
pub fn NewRect[T: Copy](min: pkg_vec.Vec2[T], max: pkg_vec.Vec2[T]) @Rect[T] {
    return alloc Rect { min: min, max: max }
}
