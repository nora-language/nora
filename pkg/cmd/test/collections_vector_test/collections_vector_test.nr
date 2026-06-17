package main

import "collections"
import "io"

fn print_val(x: #i32) {
    io.PrintLn("  element: ${x}")
}

fn is_even(x: #i32) bool {
    return x % 2 == 0
}

fn double_val(x: #i32) i32 {
    return x * 2
}

fn main() {
    io.PrintLn("--- Collections Vector Test ---")

    var v = collections.NewVector[i32](2)
    io.PrintLn("Initial len: ${v.Len()}")

    v.Push(10)
    v.Push(20)
    v.Push(30) // Resizes dynamically

    io.PrintLn("Len after 3 pushes: ${v.Len()}")
    io.PrintLn("v[0] = ${v.Get(0)}")
    io.PrintLn("v[1] = ${v.Get(1)}")
    io.PrintLn("v[2] = ${v.Get(2)}")

    v.Set(1, 99)
    io.PrintLn("v[1] after set: ${v.Get(1)}")

    var last = v.Pop()
    io.PrintLn("Popped: ${last}")
    io.PrintLn("Final len: ${v.Len()}")

    io.PrintLn("\n--- Advanced Operations Test ---")
    v.Insert(1, 50) // [10, 50, 99]
    io.PrintLn("After Insert(1, 50) len: ${v.Len()}")
    io.PrintLn("v[0]: ${v.Get(0)}, v[1]: ${v.Get(1)}, v[2]: ${v.Get(2)}")

    var removed = v.Remove(1) // [10, 99]
    io.PrintLn("Removed value: ${removed}")
    io.PrintLn("After Remove(1) len: ${v.Len()}")
    io.PrintLn("v[0]: ${v.Get(0)}, v[1]: ${v.Get(1)}")

    v.Resize(4)
    io.PrintLn("After Resize(4) len: ${v.Len()}")
    io.PrintLn("v[2] (zero-initialized): ${v.Get(2)}")

    v.Resize(2)
    io.PrintLn("After Resize(2) len: ${v.Len()}")
    io.PrintLn("v[0]: ${v.Get(0)}, v[1]: ${v.Get(1)}")

    io.PrintLn("\n--- Functional Primitives Test ---")
    // Push some more items
    v.Push(4)
    v.Push(7)
    v.Push(12) // [10, 99, 4, 7, 12]

    io.PrintLn("ForEach call:")
    v.ForEach(print_val)

    io.PrintLn("Filter call (even numbers):")
    var filtered = v.Filter(is_even)
    filtered.ForEach(print_val)

    io.PrintLn("Map call (doubled values):")
    var doubled = filtered.Map(double_val)
    doubled.ForEach(print_val)

    v.Clear()
    io.PrintLn("After Clear len: ${v.Len()}")

    io.PrintLn("\n--- String Vector Test ---")
    var sv = collections.NewVector[str](2)
    sv.Push("Hello")
    sv.Push("Nora")
    sv.Push("Standard Library")

    var i = 0
    while i < sv.Len() {
        io.PrintLn("sv[${i}] = ${sv.Get(i)}")
        i++
    }

    io.PrintLn("Collections Vector tests finished successfully!")
}
