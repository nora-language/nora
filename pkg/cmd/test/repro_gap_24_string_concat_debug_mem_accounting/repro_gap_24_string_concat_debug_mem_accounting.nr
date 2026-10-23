// Gap 4.1: std/runtime/string.c (nr_str_concat / nr_str_from_cstring / nr_strdup) allocates len+1
// bytes (adding len+1 to g_total_allocations), then sets h->count = len, h->elem_size = 1.
// When freed, nr_free only subtracts len, leaving 1 phantom leaked byte per freed string in g_total_allocations!
package main

import "io"
import "os"

extern fn get_total_allocations() i64

fn make_and_drop_str() {
    var a = "hello"
    var b = " world"
    var c = a + b
}

fn main() i32 {
    var before = get_total_allocations()
    make_and_drop_str()
    var after = get_total_allocations()
    if after != before {
        io.PrintLn("FAIL: g_total_allocations drift after freeing concat string: before=${before}, after=${after}")
        os.Exit(1)
    }
    return 0
}
