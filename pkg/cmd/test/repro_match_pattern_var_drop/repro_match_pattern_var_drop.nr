// repro_match_pattern_var_drop
//
// BUG: solver.go analyzePattern() uses s.SemanticInfo.Defs[id] to look up
// pattern-bound variable symbols. If the semantic analyzer registers match-arm
// binding variables under SemanticInfo.Uses instead of SemanticInfo.Defs
// (which can happen for certain pattern binding sites), analyzePattern returns
// nil for the symbol, never adds the variable to `visible`, and no lifecycle
// is tracked. The variable's RAII drop is never scheduled.
//
// This test binds an owned struct inside a match arm and relies on its drop
// method being called exactly once when it goes out of scope.
//
// Expected: "Item dropped: found_item" printed exactly once.
// Bug present: drop method is never called → memory leaked for pattern variables.

package main

import "io"
import "os"

type Item = struct {
    name: str
}

fn (i: #Item) drop() {
    io.PrintLn("Item dropped:", i.name)
}

type Container = enum {
    Full(item: Item),
    Empty
}

fn extract(c: Container) {
    match c {
        Full(item) => {
            // `item` is an owned Item bound by the match pattern.
            // analyzePattern must register it so it gets dropped here
            // at the end of this arm block.
            io.PrintLn("Found item:", item.name)
            // item should be dropped here (end of arm block)
        }
        Empty => {
            io.PrintLn("Nothing to extract")
        }
    }
}

fn main() i32 {
    var c1 = Full(Item { name: "found_item" })
    extract(@c1)

    var c2 = Empty
    extract(@c2)

    io.PrintLn("PASS: match pattern drop test ran")
    return 0
}
