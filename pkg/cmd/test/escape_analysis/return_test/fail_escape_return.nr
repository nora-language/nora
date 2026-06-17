package main

type RefStruct = struct {
    val: #i32
}

fn bad_return() fn() void {
    var x = 999
    var sys = RefStruct { val: #x }
    var w = fn() void {
        var y: #i32 = sys.val
    }
    return w
}

fn main() i32 {
    return 0
}
