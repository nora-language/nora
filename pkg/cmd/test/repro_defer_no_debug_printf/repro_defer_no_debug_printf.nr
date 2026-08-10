// repro_defer_no_debug_printf
//
// BUG: hir/lower.go lowerStatement() DeferStatement case has a leftover
// debug fmt.Printf call:
//
//   case *ast.DeferStatement:
//       fmt.Printf("[DEBUG] Encountered DeferStatement: %v\n", s.Call)
//       l.activeDefers = append(l.activeDefers, s.Call)
//
// This Printf fires to STDOUT during compilation (Go code), and also
// potentially during test binary execution if the compiler is invoked inline.
// More critically, it is noise that could break tooling parsing compiler output.
//
// This test simply exercises defer to confirm it compiles cleanly.
// Expected: compiles and runs, printing only the expected messages.
// Bug present: spurious "[DEBUG] Encountered DeferStatement:" lines pollute output.

package main

import "io"

fn do_work() {
    defer io.PrintLn("cleanup ran")
    io.PrintLn("doing work")
}

fn main() i32 {
    do_work()
    io.PrintLn("PASS: defer test ran without debug noise")
    return 0
}
