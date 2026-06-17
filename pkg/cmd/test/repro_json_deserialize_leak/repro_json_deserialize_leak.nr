package main

import "io"
import "serialize"

[serialize]
type TestSnapshot = struct {
    ids: i64[]
}

fn main() i32 {
    var item_str = "{\"ids\":[1, 2, 3]}"
    var snap = serialize.FromJSON[TestSnapshot](item_str)
    var sum = 0
    var i = 0
    while i < len(snap.ids) {
        sum = sum + i32(snap.ids[i])
        i = i + 1
    }
    io.PrintLn("Sum: ${sum}")
    return 0
}
