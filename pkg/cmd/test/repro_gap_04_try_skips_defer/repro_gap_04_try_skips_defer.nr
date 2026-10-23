// Gap 1.3a: TryExpression (?) early return in genTryExpression (pkg/codegen/expressions.go:2900)
// never executes registered defer statements.
package main

import "io"
import "os"

var g_defer_ran: i32 = 0

fn mark_deferred() {
    g_defer_ran = g_defer_ran + 1
}

fn always_err() Result[i32, str] {
    return Err[i32, str]("boom")
}

fn do_work() Result[i32, str] {
    defer mark_deferred()
    var x = always_err()?
    return Ok[i32, str](x)
}

fn main() i32 {
    var res = do_work()
    if g_defer_ran != 1 {
        io.PrintLn("FAIL: defer was skipped on ? early return (g_defer_ran = ${g_defer_ran})")
        os.Exit(1)
    }
    return 0
}
