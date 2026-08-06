package main

import "io"
import "math"
import "os"

fn main() i32 {
    var val32 = f32(-2.5)
    var bits32 = math.Float32bits(val32)
    var restored32 = math.Uint32bitsToFloat(bits32)

    var val64 = f64(-123.456)
    var bits64 = math.Float64bits(val64)
    var restored64 = math.Uint64bitsToFloat(bits64)

    if restored32 != val32 || restored64 != val64 {
        io.PrintLn("Error: Float bitcast restoration failed")
        os.Exit(1)
    }

    io.PrintLn("Bitcast float int test passed: bits32=", bits32, "bits64=", bits64)
    return 0
}
