// repro_try_drop_in_for_loop
//
// BUG: solver.go registerTryDrops() iterates `allVisible` to schedule drops
// for the early-exit path of the `?` operator. However, the iterator temporary
// object allocated just before a `for item in collection` loop (the struct
// stored in an HIR temp) may not be in `allVisible` at the moment
// registerTryDrops is called inside the loop body.
//
// If the iterator temp is not in `allVisible`, it is never added to TryDrops,
// so when `?` triggers an early Err return the iterator object leaks.
//
// This test verifies that all owned resources (including iterator-adjacent
// allocations) are properly cleaned up when `?` forces an early return from
// inside a for-range loop body.
//
// Expected: compiles, runs, exits cleanly, no memory leak reported.
// Bug present: iterator-scope Resource leaks on the early-return path.

package main

import "io"

type Resource = struct {
    name: str
}

fn (r: #Resource) drop() {
    io.PrintLn("Resource dropped:", r.name)
}

fn might_fail(x: i32) Result[i32, str] {
    if x >= 3 {
        return Err[i32, str]("too large")
    }
    return Ok[i32, str](x * 10)
}

// Uses `?` inside a for-range loop.
// When x==3, `?` fires and all in-scope resources must be cleaned up.
fn process() Result[i32, str] {
    var r1 = alloc Resource { name: "outer_resource" }
    var total = 0

    for i in 0..5 {
        var r2 = alloc Resource { name: "loop_resource" }
        var val = might_fail(i)?    // early exit when i >= 3
        total = total + val
    }

    return Ok[i32, str](total)
}

fn main() i32 {
    var result = process()
    match result {
        Err(e) => {
            io.PrintLn("Got expected error:", e)
        }
        Ok(v) => {
            io.PrintLn("Got value:", v)
        }
        _ => {}
    }
    io.PrintLn("PASS: try-in-for-loop drop test ran")
    return 0
}
