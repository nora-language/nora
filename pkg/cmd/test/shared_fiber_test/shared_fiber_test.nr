package main

[shared]
type SharedStruct = struct {
    value: i32
}

fn worker(wg: #SharedStruct) {
    var v = wg.value
}

fn main() i32 {
    var wg = SharedStruct {
        value: 10
    }
    scope {
        spawn worker(#wg)
        return 0

    }
}
