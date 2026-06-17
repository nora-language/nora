package main

import "io"

fn main() {
    // 1. Widening Casts
    var a_i32: i32 = 42
    var a_i64 = i64(a_i32)
    io.PrintLn("i32 to i64:", a_i64)

    var a_u16: u16 = 100
    var a_u32 = u32(a_u16)
    io.PrintLn("u16 to u32:", a_u32)

    // 2. Narrowing Casts
    var b_i64: i64 = 999999
    var b_i32 = i32(b_i64)
    io.PrintLn("i64 to i32:", b_i32)

    var b_u32: u32 = 250
    var b_u8 = u8(b_u32)
    io.PrintLn("u32 to u8:", b_u8)

    var b_i32_to_i16 = i16(b_i32)
    io.PrintLn("i32 to i16:", b_i32_to_i16)

    // 3. Signed <-> Unsigned Casts
    var c_i32: i32 = -5
    var c_u32 = u32(c_i32)
    // -5 in 32-bit unsigned representation is 4294967291
    io.PrintLn("i32 to u32 (negative):", c_u32)

    var c_u8: u8 = 255
    var c_i8 = i8(c_u8)
    // 255 in 8-bit signed representation is -1
    io.PrintLn("u8 to i8 (overflow):", c_i8)

    // 4. Float <-> Int Casts
    var d_f64: f64 = 3.85
    var d_i32 = i32(d_f64)
    io.PrintLn("f64 to i32 (truncation):", d_i32)

    var d_i32_val: i32 = 123
    var d_f64_val = f64(d_i32_val)
    io.PrintLn("i32 to f64:", d_f64_val)

    // 5. Bool <-> Int Casts
    var e_bool = true
    var e_i32 = i32(e_bool)
    io.PrintLn("bool to i32:", e_i32)

    var e_i32_val: i32 = 0
    var e_bool_val = bool(e_i32_val)
    io.PrintLn("i32 to bool (0):", e_bool_val)

    var e_i32_val2: i32 = 1
    var e_bool_val2 = bool(e_i32_val2)
    io.PrintLn("i32 to bool (1):", e_bool_val2)
}
