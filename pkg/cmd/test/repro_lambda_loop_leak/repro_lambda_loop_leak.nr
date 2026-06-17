package main

import "io"
import "serialize"

[serialize]
type TestSnapshot = struct {
    ids: i64[]
}

fn test_lambda() {
    var deserialize_fn = fn() {
        var items = alloc str[1]
        items[0] = "{\"ids\":[1, 2, 3]}"
        var i = 0
        while i < len(items) {
            var item_str = items[i]
            var snap = serialize.FromJSON[TestSnapshot](item_str)
            var sum = 0
            var j = 0
            while j < len(#snap.ids) {
                sum = sum + i32((#snap.ids)[j])
                j = j + 1
            }
            io.PrintLn("Sum: ${sum}")
            i = i + 1
        }
    }
    deserialize_fn()
}

fn main() i32 {
    test_lambda()
    return 0
}
