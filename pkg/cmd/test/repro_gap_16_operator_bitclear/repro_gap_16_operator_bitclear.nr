// Gap 2.4a: Bit-clear operator (&^ and &^=) is lexed (pkg/token/token.go:83,85),
// but &^= is missing from parser infix registration, &^ is rejected by semantic analysis,
// and codegen emits invalid C11 `&^` instead of `& ~`.
package main

import "os"

fn main() i32 {
    var x: i32 = 7
    var y: i32 = x &^ 2
    if y != 5 {
        os.Exit(1)
    }
    y &^= 1
    if y != 4 {
        os.Exit(1)
    }
    return 0
}
