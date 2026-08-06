package main

import "io"

type Heavy = struct {
    items: @str[]
}

type Payload = enum {
    Data(h: Heavy),
    None
}

fn inspect_data(p: #Payload) {
    match p {
        Data(h) => {
            io.PrintLn("First item:", h.items[0])
        }
        None => {}
    }
}

fn main() i32 {
    var arr = alloc str[2]
    arr[0] = "hello"
    arr[1] = "world"

    var p = Data(Heavy { items: @arr })
    
    // Inspect borrowed payload
    inspect_data(#p)
    inspect_data(#p)

    return 0
}
