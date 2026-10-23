// Gap 2.1c: substituteType and hasGeneric (pkg/semantic/analyzer.go:5912, 8233) omit *types.ChanType,
// leaving `chan[T]` fields in generic structs unsubstituted.
package main

import "os"

type Pipe[T] = struct {
    ch: chan[T]
}

fn main() i32 {
    var p = alloc Pipe[i32] {
        ch: make(chan[i32], 1)
    }
    p.ch <- 55
    var v = <-p.ch
    if v != 55 {
        os.Exit(1)
    }
    return 0
}
