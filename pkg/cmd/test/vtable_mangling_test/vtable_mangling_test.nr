package main

import "io"

type DummyStruct = struct {
    x: i32
}

type ComponentBox[T] = struct {
    val: T
}

fn AddDeferred[T](comp: T) any {
    var val = comp
    var box = alloc ComponentBox[T] { val: @val }
    var val_any: any = &box.val
    return val_any
}

fn main() i32 {
    var a = AddDeferred[DummyStruct](DummyStruct { x: 123 })
    io.PrintLn("Tests run successfully!")
    return 0
}
