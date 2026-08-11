// repro_lambda_defer
//
// This test ensures that hir/lower.go lowerLambdaFunction() emits defers at the end of a
// lambda that has no explicit return.
//
// Expected: "lambda defer executed" printed before "after lambda call".

package main

import "io"
import "os"

type Guard = struct {
    label: str
}

fn (g: #Guard) drop() {
    io.PrintLn("Guard dropped:", g.label)
}

fn make_executor(label: str) fn() {
    // Capture `label` by value in the lambda.
    // The defer inside must run when the lambda returns.
    return fn() {
        var g = alloc Guard { label: label }
        defer io.PrintLn("lambda defer executed for lambda")
        io.PrintLn("lambda body running for lambda")
        // guard g dropped here (RAII)
        // defer must also fire here
    }
}

fn main() i32 {
    var exec = make_executor("test_lambda")

    io.PrintLn("calling lambda...")
    exec()
    io.PrintLn("after lambda call")

    // If the bug is present, "lambda defer executed for: test_lambda" line
    // is completely absent from the output.
    io.PrintLn("PASS: lambda defer test completed (check output for defer line)")
    return 0
}
