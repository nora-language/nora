package main

import "io"
import "os"

fn main() i32 {
    io.PrintLn("Testing Numeric Shift Promotion & Large Negative Literals")

    // 1. Bitwise shift contextual literal promotion
    var base: u64 = 1
    var shifted = base << 32 // '32' is promoted to u64
    if shifted != 4294967296 {
        io.PrintLn("Error: Shift left promotion failed")
        os.Exit(1)
    }

    var val: u64 = 0xFF00000000
    var unshifted = val >> 32 // '32' is promoted to u64
    if unshifted != 0xFF {
        io.PrintLn("Error: Shift right promotion failed")
        os.Exit(1)
    }

    // 2. Negative 64-bit integer literal promotion
    var neg_large: i64 = -6364136223846793005
    if neg_large >= 0 {
        io.PrintLn("Error: Negative i64 literal promotion failed")
        os.Exit(1)
    }

    io.PrintLn("Numeric shift promotion test passed successfully!")
    return 0
}
