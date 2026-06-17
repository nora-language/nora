package main

import "io"

// 1. Identifiers (Unicode support)
var αβ = "Unicode Identifier"
var _x9 = 100
var ThisVariableIsExported = true

type CustomOption = enum {
    CustomSome(val: i32),
    CustomNone
}

// 2. Operators and punctuation
fn test_operators() {
    var a = 10
    a = a + 5
    a = a - 2
    a = a * 2
    a = a / 2
    a = a % 3

    var b = 0b1010
    b = b & 0b1100
    b = b | 0b0001
    b = b ^ 0b1111
    b = b << 1
    b = b >> 1

    // Test match and arrow
    var opt = CustomSome(a)
    match opt {
        CustomSome(val) => {
            io.PrintLn("one")
        }
        _ => {
            io.PrintLn("other")
        }
    }
}

fn main() {
    io.PrintLn(αβ)

    // 3. Integer literals
    var dec = 1234
    var hex = 0xFF00AA
    var bin = 0b1101_0010
    var oct = 0o755
    var with_underscores = 1_000_000

    io.PrintLn("Integers:", dec, hex, bin, oct, with_underscores)

    // Suffixes and Constructor-style conversions
    var small_val: i8 = i8(10)
    var large_val: u64 = u64(100_000)
    io.PrintLn("Cast Integers:", small_val, large_val)

    // 4. Floating point literals
    var f1 = 3.14159
    var f2 = 1.23e10
    var f3 = 5.0E-3
    var hex_float = 0x1.fp3

    io.PrintLn("Floats:", f1, f2, f3, hex_float)

    var f_cast = f32(9.806_65)
    var big_cast = f64(1_000.000_001)
    io.PrintLn("Cast Floats:", f_cast, big_cast)

    // 5. Imaginary literals (semantic analyzer does not support complex types yet, so we just declare them)
    var img1 = 5i
    var img2 = 3.14i
    var img3 = 1.2e-5j
    var img4 = 1_000_000i
    io.PrintLn("Imaginary:", img1, img2, img3, img4)

    test_operators()
}
