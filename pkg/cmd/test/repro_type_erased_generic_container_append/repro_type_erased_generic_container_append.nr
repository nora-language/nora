package main

import "io"

type User = struct {
    name: str
}

type Container[T] = struct {
    items: T[]
}

fn NewContainer[T]() @Container[T] {
    var list = alloc T[0]
    return alloc Container[T] {
        items: @list
    }
}

fn (c: &Container[T]) Add(item: T) {
    c.items = append(c.items, item)
}

fn main() i32 {
    var c1 = NewContainer[#str]()
    var s1 = "hello"
    var s2 = "world"
    c1.Add(#s1)
    c1.Add(#s2)

    io.PrintLn("Container 1 count:", len(c1.items))

    var u1 = User { name: "alice" }
    var u2 = User { name: "bob" }
    var c2 = NewContainer[#User]()
    c2.Add(#u1)
    c2.Add(#u2)

    io.PrintLn("Container 2 count:", len(c2.items))

    if (len(c1.items) != 2 || len(c2.items) != 2) {
        return 1
    }
    return 0
}
