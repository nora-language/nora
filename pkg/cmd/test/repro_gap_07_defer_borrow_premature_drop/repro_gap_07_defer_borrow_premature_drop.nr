// Gap 1.4b: DeferStatement is not tracked by pkg/topology/solver.go, so a local variable
// borrowed in defer is dropped after its last normal lexical use, BEFORE defer runs at scope exit.
package main

import "io"
import "os"

var g_alive: i32 = 0
var g_observed_in_defer: i32 = 0

type Guard = struct {
    id: i32
}

fn (g: #Guard) drop() {
    g_alive = 0
}

fn inspect_guard(g: #Guard) {
    g_observed_in_defer = g_alive
}

fn run_scope() {
    var g = alloc Guard { id: 7 }
    g_alive = 1
    defer inspect_guard(#g)

    // Last normal lexical use of g is here:
    var id_copy = g.id
    // solver.go drops g right here because it didn't see defer inspect_guard(#g)!
    var other = id_copy + 10
}

fn main() i32 {
    run_scope()
    if g_observed_in_defer != 1 {
        io.PrintLn("FAIL: Guard was already dropped before defer executed!")
        os.Exit(1)
    }
    return 0
}
