package main

import "io"

type Resource = struct {
    name: @str
}

fn (r: #Resource) drop() {
    io.PrintLn("Resource drop:", r.name)
}

type Item = enum {
    WithResource(res: Resource),
    Empty
}

fn inspect_item(item: #Item) {
    match item {
        WithResource(res) => {
            // 'res' is bound from borrowed 'item' (#Item)
            // It should be borrowed (#Resource), NOT moved (@Resource).
            io.PrintLn("Inspecting:", res.name)
        }
        Empty => {}
    }
}

fn main() i32 {
    var item = WithResource(Resource { name: "shared_resource" })
    
    // Inspect borrowed item twice
    inspect_item(#item)
    inspect_item(#item)

    io.PrintLn("Done inspect")
    return 0
}
