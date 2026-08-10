// repro_lambda_defer
//
// BUG: hir/lower.go lowerLambdaFunction() emits defers at the end of a
// lambda that has no explicit return, but forgets to add the lowered
// instruction to the HIR block:
//
//   // lowerLambdaFunction (BUGGY):
//   for i := len(l.activeDefers) - 1; i >= 0; i-- {
//       l.lowerExpression(l.activeDefers[i])   // ← result discarded!
//   }
//
//   // lowerFunction (CORRECT):
//   op := l.lowerExpression(l.activeDefers[i])
//   if instOp, ok := op.(*InstOperand); ok {
//       l.CurrentBlock.AddInst(instOp.Inst)    // ← properly appended
//   }
//
// Because the result of lowerExpression is thrown away, the defer call is
// completely absent from the generated C code. Defers inside lambdas silently
// do nothing.
//
// Expected: "lambda defer executed" printed before "after lambda call".
// Bug present: "lambda defer executed" never prints (defer silently dropped).

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
        defer io.PrintLn("lambda defer executed for:", label)
        io.PrintLn("lambda body running for:", label)
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
