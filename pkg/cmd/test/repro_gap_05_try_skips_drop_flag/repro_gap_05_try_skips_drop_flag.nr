// Gap 1.3b: TryExpression (?) early return in genTryExpression (pkg/codegen/expressions.go:2902)
// ignores dynamic drop flags (nr_df_...), dropping a conditionally moved variable a second time.
package main

import "io"
import "os"

var g_item_drops: i32 = 0

type Item = struct {
    v: i32
}

fn (it: #Item) drop() {
    g_item_drops = g_item_drops + 1
}

fn consume_item(it: @Item) {
    // it dropped here at end of consume_item
}

fn fail_step() Result[i32, str] {
    return Err[i32, str]("fail")
}

fn conditional_move_then_try(do_move: bool) Result[i32, str] {
    var it = alloc Item { v: 10 }
    if do_move {
        consume_item(@it)
    }
    var x = fail_step()?
    return Ok[i32, str](x)
}

fn main() i32 {
    var res = conditional_move_then_try(true)
    if g_item_drops != 1 {
        io.PrintLn("FAIL: Item dropped ${g_item_drops} times (expected 1)")
        os.Exit(1)
    }
    return 0
}
