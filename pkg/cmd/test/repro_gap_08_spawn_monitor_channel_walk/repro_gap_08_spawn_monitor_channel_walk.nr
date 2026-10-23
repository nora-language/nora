// Gap 1.4c: ast.Inspect (pkg/parser/ast/walk.go:175) omits SpawnExpression.MonitorChannel,
// so the lease solver never sees mon_ch used in spawn(mon_ch) and drops mon_ch before spawn!
package main

import "io"

fn worker() {
}

fn main() i32 {
    var mon_ch = make(chan[str], 1)
    spawn(mon_ch) worker()
    var msg = <-mon_ch
    return 0
}
