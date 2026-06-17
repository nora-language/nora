package main

type Entry = struct {
    x: i32
}

fn main() i32 {
    var m = Map[str, i32]()
    var key = Entry {
        x: 42
    }

    // 1. Should fail because map key type is str, not Entry struct
    var val = m[key]

    var arr = List[i32]()
    // 2. Should fail because index must be an integer, not str
    var item = arr["first"]

    return 0
}
