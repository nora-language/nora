package main

import "io"

fn array() {
    var b = [1, 2, 3]
    io.PrintLn("Initial length:", len(b))
    for i, a in b {
        io.PrintLn("${i} ${a}")
    }

    var ind = 0
    while ind < len(b) {
        var r = b[ind]
        io.PrintLn("${ind + 1} ${r}")
        ind = ind + 1
    }

    for x in 0 .. len(b) {
        var r = b[x]
        io.PrintLn("${x + 1} ${r}")
    }

    for x in 0 .. len(b) {
        io.PrintLn("${x + 1} aho ${b[x]}")
    }

    {
        var arr = alloc i32[5]
        io.PrintLn("Array length: ", len(arr))

        arr[0] = 10
        arr[4] = 50

        var r = arr[0]
        var q = arr[4]

        io.PrintLn("arr[0] = ${r}  arr[4] = ${q}")
    }
    {
        var list = make(List[i32], 5)
        list[0] = 42
        io.PrintLn("list[0]: ${list[0]}")
    }
}
