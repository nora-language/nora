// Gap 2.1b: cloneAndSubstituteHelper (pkg/semantic/analyzer.go:6966) omits *ast.ChanType,
// so `make(chan[T], 1)` inside a generic function never substitutes T.
package main

import "io"
import "os"

fn roundtrip_chan[T](val: T) T {
    var ch = make(chan[T], 1)
    ch <- val
    var out = <-ch
    return out
}

fn main() i32 {
    var v = roundtrip_chan[i32](77)
    if v != 77 {
        io.PrintLn("FAIL: expected 77, got ${v}")
        os.Exit(1)
    }
    return 0
}
