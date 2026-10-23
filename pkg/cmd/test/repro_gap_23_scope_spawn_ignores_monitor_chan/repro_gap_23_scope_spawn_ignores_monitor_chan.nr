// Gap 3.1: ScopeExpression bypasses HIR lowering via ASTExpr, so `spawn(mon_ch) worker()` inside
// `scope { ... }` is compiled via legacy genSpawnExpression (pkg/codegen/expressions.go:2165),
// which completely ignores `e.MonitorChannel` (passing NULL as monitor_chan to nr_spawn)!
package main

import "os"

fn worker() {
    panic("boom")
}

fn main() i32 {
    var mon_ch = make(chan[str], 1)
    scope {
        spawn(mon_ch) worker()
    }
    var status = <-mon_ch
    if status != "boom" {
        os.Exit(1)
    }
    return 0
}
