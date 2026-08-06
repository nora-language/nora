package main

import "io"

type Item = struct {
    name: str
}

fn (i: #Item) drop() {
    io.PrintLn("Item dropped:", i.name)
}

fn get_err() Result[i32, str] {
    return Err[i32, str]("failure")
}

fn process_with_try() Result[i32, str] {
    var item1 = Item { name: "item1" }
    var item2 = Item { name: "item2" }

    // '?' fails here and returns early with Err[i32, str]("failure")
    // item1 and item2 MUST be dropped before returning!
    var x = get_err()?

    var item3 = Item { name: "item3" }
    return Ok[i32, str](x)
}

fn main() i32 {
    var res = process_with_try()
    match res {
        Ok(v) => {
            io.PrintLn("Success:", v)
        }
        Err(e) => {
            io.PrintLn("Caught expected error:", e)
        }
        _ => {}
    }
    return 0
}
