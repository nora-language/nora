package main

import "io"

type Vector[T] = struct {
    data: @T[],
    size: i32,
    capacity: i32
}

fn new_vector[T](cap: i32) @Vector[T] {
    var c = cap
    if (c < 1) {
        c = 1
    }
    return alloc Vector[T] {
        data: alloc T[c],
        size: 0,
        capacity: c
    }
}

fn (v: #Vector[T]) len[T]() i32 {
    return v.size
}

fn (v: &Vector[T]) push[T](val: T) {
    if (v.size == v.capacity) {
        var new_cap = v.capacity * 2
        var new_data = alloc T[new_cap]

        // io.PrintLn("Resizing vector: ${v.capacity} -> ${new_cap}")
        var i = 0
        while i < v.size {
            new_data[i] = @v.data[i]
            i++
        }

        v.data = @new_data
        v.capacity = new_cap
    }

    v.data[v.size] = val
    v.size = v.size + 1
}

fn (v: #Vector[T]) get[T](index: i32) T {
    if (index < 0 || index >= v.size) {
        io.PrintLn("Error: Index ${index} out of bounds (size ${v.size})")
    }
    return v.data[index]
}

fn (v: &Vector[T]) pop[T]() T {
    if (v.size == 0) {
        io.PrintLn("Error: Pop from empty vector")
    }
    v.size = v.size - 1
    return v.data[v.size]
}

fn (v: &Vector[T]) set[T](index: i32, val: T) {
    if (index >= 0 && index < v.size) {
        v.data[index] = val
    }
}

fn main() i32 {
    io.PrintLn("--- Generic Vector Test ---")

    var v = new_vector[i32](2)
    io.PrintLn("Initial len: ${v.len()}")

    v.push(10)
    v.push(20)
    v.push(30) // Should resize

    io.PrintLn("Len after 3 pushes: ${v.len()}")
    io.PrintLn("v[0] = ${v.get(0)}")
    io.PrintLn("v[1] = ${v.get(1)}")
    io.PrintLn("v[2] = ${v.get(2)}")

    v.set(1, 99)
    io.PrintLn("v[1] after set: ${v.get(1)}")

    var last = v.pop()
    io.PrintLn("Popped: ${last}")
    io.PrintLn("Final len: ${v.len()}")

    io.PrintLn("\n--- String Vector Test ---")
    var sv = new_vector[str](2)
    sv.push("Hello")
    sv.push("Nora")
    sv.push("Generics")

    var i = 0
    while i < sv.len() {
        io.PrintLn("sv[${i}] = ${sv.get(i)}")
        i++
    }

    io.PrintLn("Vector tests finished successfully!")
    return 0
}
