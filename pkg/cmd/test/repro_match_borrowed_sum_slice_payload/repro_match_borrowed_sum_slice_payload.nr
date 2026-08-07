package main

import "io"

type DataPayload = enum {
    Values(arr: @str[]),
    Empty
}

fn inspect_payload(p: #DataPayload) i32 {
    match p {
        Values(items) => {
            io.PrintLn("First item:", items[0])
            return 1
        }
        Empty => {
            return 0
        }
    }
}

fn main() i32 {
    var list = alloc str[2]
    list[0] = "hello"
    list[1] = "world"

    var payload = Values(@list)

    // Inspect borrowed sum type twice without consuming it
    var r1 = inspect_payload(#payload)
    var r2 = inspect_payload(#payload)

    io.PrintLn("r1:", r1, "r2:", r2)
    if (r1 != 1 || r2 != 1) {
        return 1
    }
    return 0
}
