// Gap 1.4a: SelectStatement is missing from pkg/topology/solver.go's control-flow block analysis
// and pkg/topology/cfg.go, so moving an owned variable inside a select case is not tracked as a
// conditional move, causing the variable to be dropped a second time at function exit!
package main

import "io"
import "os"

var g_tok_drops: i32 = 0

type Token = struct {
    id: i32
}

fn (t: #Token) drop() {
    g_tok_drops = g_tok_drops + 1
}

fn consume(t: @Token) {
}

fn main() i32 {
    var tok = alloc Token { id: 1 }
    var ch = make(chan[i32], 1)
    ch <- 10

    select {
        case var x = <-ch: {
            consume(@tok)
        }
    }
    if g_tok_drops != 1 {
        io.PrintLn("FAIL: expected Token to be dropped once, dropped ${g_tok_drops} times")
        os.Exit(1)
    }
    return 0
}
