// Gap 1.2b: AnalyzeLambdaExpression (pkg/semantic/analyzer.go:4820-4829) maps '#' to LeaseWrite
// instead of LeaseRead when building fnType.ParamLeases.
package main

import "io"
import "os"

fn main() i32 {
    var read_len = fn(s: #str) i32 {
        return len(s)
    }
    var msg = "hello"
    var n = read_len(#msg)
    if n != 5 {
        io.PrintLn("FAIL: expected 5")
        os.Exit(1)
    }
    return 0
}
