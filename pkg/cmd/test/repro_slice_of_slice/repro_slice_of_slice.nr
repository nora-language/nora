package main

import "io"
import "os"

fn main() i32 {
    var arr = alloc i32[10]
    for i in 0..10 {
        arr[i] = (i + 1) * 10
    }

    var slice1 = arr[2..8] // [30, 40, 50, 60, 70, 80]
    var slice2 = slice1[1..4] // [40, 50, 60]

    if len(slice2) != 3 || slice2[0] != 40 || slice2[2] != 60 {
        io.PrintLn("Error: Slice of slice values wrong")
        os.Exit(1)
    }

    io.PrintLn("Slice of slice indexing test passed: len=", len(slice2), "val0=", slice2[0], "val2=", slice2[2])
    return 0
}
