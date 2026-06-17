package main

import "io"

fn main() i32 {
    io.PrintLn("=== Nora Numeric Promotion & Coercion Test ===")

    // 1. Large values (> 2^31 - 1) should automatically promote to i64
    var large_val = 6364136223846793005 // Fits in i64, automatically typed as i64
    io.PrintLn("Large Val (automatically i64):", large_val)

    // 2. Contextual type coercion in infix expressions
    var my_i64: i64 = 1000
    var result_div = my_i64 / 10 // '10' is default i32, automatically promoted to i64 to match my_i64!
    io.PrintLn("Contextual Division result:", result_div)

    var result_mul = 10 * my_i64 // '10' is default i32, automatically promoted to i64 to match my_i64!
    io.PrintLn("Contextual Multiplication result:", result_mul)

    io.PrintLn("Nora Numeric Promotion & Coercion Test completed successfully!")
    return 0
}
