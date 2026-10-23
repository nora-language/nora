// Gap 4.2f: context.Cancel (std/context/context.nr:15-26) sends a single `true` value into
// `done_chan: make(chan[bool], 1)`. When 2+ fibers wait on `<-ctx.Done()`, only 1 fiber wakes up
// and the remaining fibers deadlock!
package main

import "context"

fn wait_for_cancel(ctx: #context.Context, out: chan[i32]) {
    var ch = ctx.Done()
    var done = <-ch
    out <- 1
}

fn main() i32 {
    var ctx = context.WithCancel()
    var out = make(chan[i32], 2)

    spawn wait_for_cancel(#ctx, out.clone())
    spawn wait_for_cancel(#ctx, out.clone())

    (&ctx).Cancel("shutdown")

    var a = <-out
    var b = <-out
    return 0
}
