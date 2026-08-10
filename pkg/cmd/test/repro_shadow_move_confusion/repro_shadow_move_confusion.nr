// repro_shadow_move_confusion
//
// BUG: solver.go isSameSelector() compares identifiers by string name (id.Value)
// instead of by semantic symbol pointer. Two variables with the same name in
// nested scopes will be conflated, causing a false "use-after-move" error on
// the outer variable after the inner variable's field is moved.
//
// Expected: compiles and prints outer label correctly.
// Bug present: compiler emits a spurious "use of moved value 'node'" error
// even though only the *inner* node.label was moved.

package main

import "io"
import "os"

type Node = struct {
    label: @str
}

fn main() i32 {
    var node = alloc Node { label: "outer" }

    scope {
        // This `node` shadows the outer one – it is a DIFFERENT symbol
        // at the same string name "node".
        var node = alloc Node { label: "inner" }

        // Move the INNER node.label only.
        var taken = @node.label
        io.PrintLn("inner label:", taken)
        // inner node (partially moved, label gone) is dropped here
    }

    // The OUTER node.label was never touched.
    // With isSameSelector bug, solver treats outer's field-move record as
    // matching the inner's move, reporting a false "use-after-move" here.
    io.PrintLn("outer label:", node.label)

    if node.label != "outer" {
        io.PrintLn("FAIL: outer label corrupted")
        os.Exit(1)
    }

    io.PrintLn("PASS: outer variable unaffected by inner shadow move")
    return 0
}
