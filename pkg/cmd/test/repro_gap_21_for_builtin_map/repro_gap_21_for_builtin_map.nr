// Gap 3.2b: `for k, v in m` where m is built-in Map[K, V] is allowed by semantic analysis
// (pkg/semantic/analyzer.go:4566), but unhandled in HIR lower.go and codegen statements.go!
package main

import "os"

fn main() i32 {
    var m: Map[str, i32] = {"a": 1, "b": 2}
    var total = 0
    for k, v in m {
        total = total + v
    }
    if total != 3 {
        os.Exit(1)
    }
    return 0
}
