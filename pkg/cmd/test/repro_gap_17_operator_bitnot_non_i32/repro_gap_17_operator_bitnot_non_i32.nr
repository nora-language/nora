// Gap 2.4b: Bitwise NOT (~) in pkg/semantic/analyzer.go:1879 only allows i32 and int,
// rejecting u8, u16, u32, u64, i8, i16, i64, and byte.
package main

import "os"

fn main() i32 {
    var a: u8 = 0
    var b: u8 = ~a
    if b != 255 {
        os.Exit(1)
    }
    var c: i64 = 0
    var d: i64 = ~c
    if d != -1 {
        os.Exit(1)
    }
    return 0
}
