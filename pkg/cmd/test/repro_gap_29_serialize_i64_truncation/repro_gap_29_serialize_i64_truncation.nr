// Gap 4.2e: nr_serialize_json_i64 (std/serialize/serialize.nr:36-38) casts i64 to i32(val),
// truncating 64-bit integers > 2^31 - 1 during JSON serialization!
package main

import "serialize"
import "io"
import "os"

fn main() i32 {
    var big: i64 = 5000000000
    var s = serialize.nr_serialize_json_i64(big)
    if s != "5000000000" {
        io.PrintLn("FAIL: expected 5000000000, got ${s}")
        os.Exit(1)
    }
    return 0
}
