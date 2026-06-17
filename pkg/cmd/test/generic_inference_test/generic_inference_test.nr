package main

import "io"

fn identity[T](x: T) T {
    return x
}

fn add[T](a: T, b: T) T {
    return a
}

type Box[T] = struct {
    value: T
}

fn get_box_value[T](b: Box[T]) T {
    return b.value
}

fn main() i32 {
    var x = identity(42)
    var y = identity("hello")

    io.PrintLn("x = ${x}")
    io.PrintLn("y = ${y}")

    var sum = add(10, 20)
    io.PrintLn("sum = ${sum}")

    var b = Box[i32] {
        value: 100
    }
    var val = get_box_value(b)
    io.PrintLn("box value = ${val}")

    return 0
}
