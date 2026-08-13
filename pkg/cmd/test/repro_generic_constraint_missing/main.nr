package main

pub type Vector3[T: Copy] = struct {
    x: T
}

pub fn (v: #Vector3[T]) Sub[T: Copy](other: #Vector3[T]) Vector3[T] {
    return Vector3[T]{ x: v.x }
}

// Bug trigger: caller lacks T: Copy constraint
pub fn TestSub[T](v: #Vector3[T]) {
    var other = Vector3[T]{ x: v.x }
    var diff = v.Sub[T](#other)
}

pub fn main() {
    var v = Vector3[f32]{ x: f32(1.0) }
    TestSub[f32](#v)
}
