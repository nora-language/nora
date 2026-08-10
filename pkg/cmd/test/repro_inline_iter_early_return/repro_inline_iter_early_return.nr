// repro_inline_iter_early_return
//
// BUG: hir/optimize/inliner.go transformReturns() recurses into HIRIf and
// HIRLoop, but falls through on IteratorLoop (the default case). When an
// `[inline]` function contains an iterator-protocol for-loop with an early
// `return`, the Ret instruction inside the IteratorLoop body is NOT converted
// to the `goto _inline_end_xxx` pattern. The resulting C code emits a bare
// `return` from within the inlined block, causing either a C compile error or
// silent wrong-return in the caller.
//
// The function uses a struct with Next() → Option[T], which the HIR lowerer
// turns into an IteratorLoop element (not the range-based HIRLoop).
//
// Expected: compiles and returns the first value over `threshold`.
// Bug present: C compile error (return inside inlined block) or wrong result.

package main

import "io"
import "os"

// A simple integer range iterator (uses Next() -> Option[i32])
type CounterIter = struct {
    current: i32
    limit: i32
}

pub fn (c: &CounterIter) Next() Option[i32] {
    if c.current >= c.limit {
        return None[i32]
    }
    var v = c.current
    c.current = c.current + 1
    return Some[i32](v)
}

// This function is inlined into the call site.
// The early `return` inside the iterator for-loop triggers the bug.
[inline]
fn first_over(limit: i32, threshold: i32) i32 {
    var it = CounterIter { current: 0, limit: limit }
    for v in it {
        if v > threshold {
            return v    // ← Ret inside IteratorLoop, not transformed by inliner
        }
    }
    return -1
}

fn main() i32 {
    var result = first_over(10, 5)
    io.PrintLn("first over 5:", result)

    if result != 6 {
        io.PrintLn("FAIL: expected 6, got:", result)
        os.Exit(1)
    }

    io.PrintLn("PASS: inline iterator early return works")
    return 0
}
