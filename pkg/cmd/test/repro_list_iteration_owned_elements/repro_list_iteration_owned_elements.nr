package main

import "io"

type Item = struct {
    name: str
}

fn main() i32 {
    var items = alloc Item[2]
    items[0] = Item { name: "apple" }
    items[1] = Item { name: "banana" }

    for item in items {
        io.PrintLn("Item name:", item.name)
    }

    return 0
}
