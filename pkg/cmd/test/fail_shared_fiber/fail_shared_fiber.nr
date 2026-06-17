package main

type NormalStruct = struct {
    value: i32
}

fn worker(wg: #NormalStruct) {
    var v = wg.value
}

fn main() i32 {
    var wg = NormalStruct {
        value: 10
    }
    scope {
        spawn worker(#wg)
        return 0

    }
}
