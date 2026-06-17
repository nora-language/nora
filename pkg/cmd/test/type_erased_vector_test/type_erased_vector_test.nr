package main

import "collections"
import "io"

type Item = struct {
    name: str
}

fn main() i32 {
    // Vector of dynamic strings
    var v_str = collections.NewVector[str](2)
    var s1 = "item_" + "1"
    var s2 = "item_" + "2"
    v_str.Push(@s1)
    v_str.Push(@s2)

    // Vector of owned structs containing strings
    var v_item = collections.NewVector[Item](2)
    var i1 = Item { name: "dynamic_" + "item1" }
    var i2 = Item { name: "dynamic_" + "item2" }
    v_item.Push(@i1)
    v_item.Push(@i2)

    io.PrintLn("Vector of strings len: ${v_str.Len()}")
    io.PrintLn("Vector of items len: ${v_item.Len()}")
    return 0
}
