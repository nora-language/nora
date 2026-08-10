// fail_while_loop_field_move
//
// BUG: solver.go WhileStatement analysis runs pass 2 on the ALREADY-MUTATED
// bodyLifecycles from pass 1 instead of a fresh clone. This is the while-loop
// counterpart to fail_partial_move_loop.
//
// `@w.data` moves the owned str field out of `w`. On the second iteration the
// field is already gone — the topological solver MUST detect this violation.
//
// Expected: compilation error "use of partially moved value 'w'"
// Bug present: compiles, then crashes or silently corrupts memory at runtime.

package main

import "io"

type Wrapper = struct {
    data: @str
}

fn main() i32 {
    var w = alloc Wrapper { data: "hello" }

    var count = 0
    while count < 2 {
        // Second iteration: w.data is already moved — solver must catch this!
        var x = @w.data
        io.PrintLn("moved:", x)
        count = count + 1
    }

    return 0
}
