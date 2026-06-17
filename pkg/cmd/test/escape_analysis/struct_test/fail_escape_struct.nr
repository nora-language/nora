package main

import "io"

type BigData = struct {
    value: i32
    padding1: i32
    padding2: i32
    padding3: i32
}

type RefWrap = struct {
    lease: #BigData
}

fn create_closure() fn() void {
    var data = BigData{value: 999, padding1: 0, padding2: 0, padding3: 0}
    var w = RefWrap{lease: #data}
    var f = fn() void {
        var val: #BigData = w.lease
        io.PrintLn("Value: ${io.FormatInt(val.value)}")
    }
    return f
}

fn smash_stack() void {
    var a = 111
    var b = 222
    var c = 333
    var d = 444
    var e = 555
    io.PrintLn("Smashing stack...")
}

fn main() i32 {
    var f = create_closure()
    smash_stack()
    f()
    return 0
}
