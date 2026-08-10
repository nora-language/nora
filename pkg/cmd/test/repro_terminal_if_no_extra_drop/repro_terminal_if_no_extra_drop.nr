// repro_terminal_if_no_extra_drop
//
// BUG: solver.go isTerminalBlock() only scans the direct children of a block.
// An if-expression whose BOTH branches end in `return` is a de-facto terminal
// block, but isTerminalBlock() misses it (it sees only the ExpressionStatement
// wrapping the IfExpression, not the returns inside). This leads to redundant
// drop registrations for variables that were already dropped inside the terminal
// branches, potentially causing double-free.
//
// Expected: compiles, runs, drops each Handle exactly once.
// Bug present: Handle may be double-dropped, producing "use-after-free" or
// duplicate "dropped:" output lines.

package main

import "io"
import "os"

type Handle = struct {
    id: i32
}

fn (h: #Handle) drop() {
    io.PrintLn("Handle dropped:", h.id)
}

// Both branches return — the function body block is effectively terminal
// after this if-expression, yet isTerminalBlock misses it.
fn pick(flag: bool) @Handle {
    var a = alloc Handle { id: 1 }
    var b = alloc Handle { id: 2 }

    if flag {
        return @a   // a returned (moved), b dropped here
    } else {
        return @b   // b returned (moved), a dropped here
    }
}

fn main() i32 {
    var h = pick(true)
    io.PrintLn("got handle:", h.id)

    if h.id != 1 {
        io.PrintLn("FAIL: wrong handle id")
        os.Exit(1)
    }

    io.PrintLn("PASS: terminal-if branch test passed")
    return 0
}
