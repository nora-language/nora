// Gap 1.1: return inside if/else emits PreDrops twice in lowerStatementsWithDrops (pkg/hir/lower.go:1561)
// causing Use-After-Free (PreDrops run BEFORE the return expression is evaluated) and Double-Free
// (when the local variable inside the if-block has no dynamic drop flag).
package main

import "io"
import "os"

var g_drops: i32 = 0
var g_alive: i32 = 0

type Resource = struct {
    val: i32
}

fn (r: #Resource) drop() {
    g_drops = g_drops + 1
    g_alive = 0
}

fn read_while_alive(r: #Resource) i32 {
    if g_alive == 0 {
        return -1
    }
    return r.val
}

fn compute(flag: bool) i32 {
    if flag {
        var r = alloc Resource { val: 42 }
        g_alive = 1
        // lowerStatementsWithDrops emits PreDrops (dropping r and setting g_alive=0) BEFORE
        // read_while_alive(#r) is evaluated, AND ReturnStatement emits PreDrops a second time!
        return read_while_alive(#r)
    }
    return 0
}

fn main() i32 {
    var v = compute(true)
    if v != 42 {
        io.PrintLn("FAIL: r was dropped BEFORE return expression was evaluated (v=${v})")
        os.Exit(1)
    }
    if g_drops != 1 {
        io.PrintLn("FAIL: expected r to be dropped once, but drop ran ${g_drops} times")
        os.Exit(1)
    }
    return 0
}
