// Gap 3.3: In HIR select codegen (pkg/codegen/hir_codegen.go:356-361), unselected send cases
// only free *types.ChanType values and leak owned list/struct payloads!
package main

import "os"

fn main() i32 {
    var unready_send = make(chan[i32[]], 0)
    var ready_recv = make(chan[i32], 1)
    ready_recv <- 42

    var got = 0
    select {
        case unready_send <- [10, 20, 30]: {
            got = -1
        }
        case var v = <-ready_recv: {
            got = v
        }
    }
    if got != 42 {
        os.Exit(1)
    }
    return 0
}
