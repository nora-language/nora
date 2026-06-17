package main

var global_closure: fn() void

fn main() i32 {
    var x = 999
    var w = fn() void {
        var y: #i32 = #x
    }
    global_closure = w
    return 0
}
