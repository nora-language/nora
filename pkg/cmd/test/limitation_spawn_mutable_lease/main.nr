package main

import "io"
import "collections"

pub fn main() {
    var vec = collections.NewVector[i32](5)
    vec.Push(10)
    vec.Push(20)
    vec.Push(30)

    // SOLVED: Cross-fiber mutable lease limitation is circumvented via ParMap.
    // ParMap[T, R] reads elements via #T (read-only borrow) and maps to a new @Vector[R].
    // For primitives (i32, f32, etc.) which are Copy types, the borrow is automatically
    // dereferenced, so the mapper receives the value directly.
    var res = vec.ParMap[i32, i32](fn(val: i32) i32 {
        return val * 2
    })

    io.PrintLn("Result[0]: ${res.Get[i32](0)}")
    io.PrintLn("Result[1]: ${res.Get[i32](1)}")
    io.PrintLn("Result[2]: ${res.Get[i32](2)}")

    _ = @res
    _ = @vec
}
