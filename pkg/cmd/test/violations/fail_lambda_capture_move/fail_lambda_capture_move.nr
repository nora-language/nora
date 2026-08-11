// fail_lambda_capture_move
//
// This test ensures that the topological solver tracks captures inside lambdas.
// Since `label` is captured by value, it is owned by the lambda closure.
// When it is moved into `Guard.label`, it is consumed.
// A subsequent use in `io.PrintLn` must trigger a "use of moved value" error.
//
// Expected: compilation fails with "use of moved value 'label'"

package main

import "io"

type Guard = struct {
    label: str
}

fn make_executor(label: str) fn() {
    return fn() {
        var g = alloc Guard { label: label } // label is moved here
        io.PrintLn("lambda body running for:", label) // use after move!
    }
}

fn main() i32 {
    return 0
}
