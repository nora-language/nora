package main

import "io"
import "./pkg_rect"
import "./pkg_vec"

// Repro: Cross-package generic return type mismatch with nested generics.
//
// Expected: pkg_rect.NewRect[i32](min, max) compiles successfully.
// Actual:   Semantic error inside pkg_rect.nr:
//   "cannot return value of type Rect from function returning pkg_rect_Rect_XXXXXXXX"
//
// Root cause: When NewRect[T] is monomorphized from main with T=i32, the compiler
// assigns the return type the cross-package prefixed name `pkg_rect_Rect_XXXXXXXX`.
// But inside NewRect's body, `Rect[T] { ... }` is still the local name `Rect`.
// The compiler cannot match these two names, causing a false type error.
//
// Note: This bug only triggers when the struct contains ANOTHER generic struct
// as a field (Vec2[T] inside Rect[T]). Structs with only primitive fields work.

pub fn main() {
    var min = pkg_vec.NewVec2[i32](0, 0)
    var max = pkg_vec.NewVec2[i32](10, 10)
    // Bug triggers here during monomorphization of NewRect:
    var r = pkg_rect.NewRect[i32](min, max)
    io.PrintLn("rect created\n")
}
