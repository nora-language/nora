package main

import "io"

type Data = struct {
    tag: i32
}

fn inspect(d: #Data) str {
    return "borrowed"
}

fn inspect(d: Data) str {
    return "owned"
}

fn main() i32 {
    var d1 = Data { tag: 1 }
    var d2 = Data { tag: 2 }

    var res1 = inspect(#d1)
    var res2 = inspect(@d2)

    io.PrintLn("res1:", res1)
    io.PrintLn("res2:", res2)

    if (res1 != "borrowed" || res2 != "owned") {
        return 1
    }
    return 0
}
