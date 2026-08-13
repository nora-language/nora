package main

import "io"

pub type Vector3[T] = struct {
    val: T
}

pub fn (v: Vector3[T]) Add(other: @Vector3[T]) @Vector3[T] {
    return alloc Vector3[T] { val: v.val }
}

pub type Triangle[T] = struct {
    a: @Vector3[T]
}

pub fn (t: Triangle[T]) GetA() @Vector3[T] {
    return alloc Vector3[T] { val: t.a.val }
}

pub type Vector[T] = struct {
    item: @T
}

pub fn (v: Vector[T]) Get() @T {
    // simplified
    return @v.item
}

pub type Mesh[T] = struct {
    triangles: @Vector[Triangle[T]]
}

pub type Shape[T] = enum {
    MeshShape(m: #Mesh[T])
}

pub fn CheckCollision[T](shape_a: #Shape[T]) {
    match shape_a {
        MeshShape(m) => {
            var tri = m.triangles.Get()
            var v3 = tri.GetA()
            var v3_2 = alloc Vector3[T] { val: v3.val }
            var res = v3.Add(@v3_2)
        }
    }
}

pub fn main() i32 {
    var v3 = alloc Vector3[f64] { val: 0.0 }
    var tri = alloc Triangle[f64] { a: @v3 }
    var vec = alloc Vector[Triangle[f64]] { item: @tri }
    var mesh = alloc Mesh[f64] { triangles: @vec }
    var shape = Shape.MeshShape[f64](#mesh)

    CheckCollision[f64](#shape)
    
    io.PrintLn("Success")
    return 0
}
