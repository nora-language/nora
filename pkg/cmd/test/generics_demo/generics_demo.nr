package main

import "io"

fn identity[T](x: T) T {
    return x
}

fn add[T](a: T, b: T) T {
    return a + b
}

fn main() {
    var a = identity[i32](42)
    var b = identity[str]("hello monomorphization")

    var c = add[i32](10, 20)
    var d = add[f64](1.5, 2.5)

    io.PrintLn("a: ${a}")
    io.PrintLn("b: ${b}")
    io.PrintLn("c: ${c}")
    io.PrintLn("d: ${d}")
}
