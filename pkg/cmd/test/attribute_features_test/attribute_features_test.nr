package attribute_features_test

import "std/io"

[native("int")]
pub type NativeVec = struct {
    v0: f32
    v1: f32
    v2: f32
    v3: f32
}

[vector_size(16)]
pub type SimdVec = struct {
    v0: f32
    v1: f32
    v2: f32
    v3: f32
}

[cfg("os=windows")]
pub fn os_specific() {
    io.PrintLn("windows")
}

[cfg("os=linux")]
pub fn os_specific() {
    io.PrintLn("linux")
}

[inline]
pub fn inline_func() {
    io.PrintLn("inline")
}

pub extern fn do_something_with_simd(a: NativeVec, b: SimdVec)

pub fn main() {
    inline_func()
    os_specific()
}
