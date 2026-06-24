package main

import "io"

pub type Vector3[T] = struct {
    x: T,
    y: T,
    z: T,
}

pub type Body[T] = struct {
    position: Vector3[T],
}

pub type BodyPtr[T] = struct {
    ptr: @Body[T],
}

pub fn (v: &BodyPtr[T]) DoMap[T](f: #fn(&BodyPtr[T])) {
    f(v)
}

pub fn Step[T](wrapper: &BodyPtr[T], gravity: &Vector3[T]) {
    wrapper.DoMap[T](fn(b_wrapper: &BodyPtr[T]) {
        var g = #gravity
        b_wrapper.ptr.position.x = g.x
    })
}

fn main() {
    var v = Vector3[f64] { x: 1.0, y: 2.0, z: 3.0 }
    var body = alloc Body[f64] { position: Vector3[f64] { x: 0.0, y: 0.0, z: 0.0 } }
    var wrapper = BodyPtr[f64] { ptr: @body }
    Step[f64](&wrapper, &v)
    io.PrintLn("Test Passed")
}
