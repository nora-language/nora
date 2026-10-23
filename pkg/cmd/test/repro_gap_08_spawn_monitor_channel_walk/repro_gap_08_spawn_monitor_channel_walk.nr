// Gap 1.4c: ast.Inspect (pkg/parser/ast/walk.go:175) omits SpawnExpression.MonitorChannel,
// so the lease solver never sees mon_ch used in spawn(mon_ch) and drops mon_ch before spawn!
package main

import "io"

fn worker() {
    panic("worker crashed")
}

fn monitor_listener(mon_ch: #chan[str], out_ch: #chan[i32]) {
    var msg = <-mon_ch
    io.PrintLn("caught:", msg)
    out_ch <- 1
}

fn main() i32 {
    var mon_ch = make(chan[str], 1)
    var out_ch = make(chan[i32], 1)
    spawn monitor_listener(#mon_ch, #out_ch)
    spawn(mon_ch) worker()
    var ok = <-out_ch
    io.PrintLn("done:", ok)
    return 0
}
