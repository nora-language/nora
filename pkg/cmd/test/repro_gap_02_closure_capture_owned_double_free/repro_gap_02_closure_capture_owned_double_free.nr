// Gap 1.2: Capturing an owned variable in a closure copies it into the closure env
// and drops it in both the closure env destructor AND the enclosing function scope (Double-Free).
package main

import "io"
import "os"

var g_box_drops: i32 = 0

type Box = struct {
    id: i32
}

fn (b: #Box) drop() {
    g_box_drops = g_box_drops + 1
}

fn run_closure() i32 {
    var b = alloc Box { id: 99 }
    var f = fn() i32 {
        return b.id
    }
    return f()
}

fn main() i32 {
    var res = run_closure()
    if res != 99 {
        io.PrintLn("FAIL: expected 99")
        os.Exit(1)
    }
    if g_box_drops != 1 {
        io.PrintLn("FAIL: expected Box to be dropped once, dropped ${g_box_drops} times")
        os.Exit(1)
    }
    return 0
}
