package main

import "io"

fn main() {
    var arr = [10, 20, 30]

    io.PrintLn("Initial length:", len(arr))
    io.PrintLn("arr[0] =", arr[0])

    io.PrintLn("Modifying arr[0] to 100...")
    arr[0] = 100

    io.PrintLn("New arr[0] =", arr[0])

    io.PrintLn("Iterating refined array:")
    for i, x in arr {
        io.PrintLn(i, ":", x)
    }
}
