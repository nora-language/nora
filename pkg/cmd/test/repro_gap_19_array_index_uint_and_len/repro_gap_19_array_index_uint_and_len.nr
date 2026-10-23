// Gap 2.5: len() rejects fixed-size arrays T[N] (pkg/semantic/analyzer.go:3714)
// and array indexing rejects u32/u8/i16 indices (isIntegerType at analyzer.go:7915).
package main

import "os"

fn main() i32 {
    var fixed: i32[3] = [10, 20, 30]
    var idx: u32 = 1
    var val = fixed[idx]
    var n = len(fixed)
    if val != 20 || n != 3 {
        os.Exit(1)
    }
    return 0
}
