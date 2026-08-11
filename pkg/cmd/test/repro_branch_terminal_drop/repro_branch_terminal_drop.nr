package main
import "io"
type Resource = struct { name: str }
fn process(items: #i32[]) i32 {
    var result = 0
    for item in @items {
        var r = alloc Resource { name: "item_resource" }
        if item > 5 {
            break
        }
        io.PrintLn(r.name)
        result = result + item
    }
    return result
}
fn main() i32 {
    var items = [1, 2, 3, 10, 4, 5]
    io.PrintLn("Sum:", process(#items))
    return 0
}
