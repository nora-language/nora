// Gap 3.2a: `for x in fixed_arr` where fixed_arr is a fixed-size array T[N] (*types.ArrayType)
// emits `array_count(arrTemp)` on a by-value C struct `nr_arr_int32_t_3`, failing C compilation!
package main

import "os"

fn main() i32 {
    var arr: i32[3] = [1, 2, 3]
    var sum = 0
    for x in arr {
        sum = sum + x
    }
    if sum != 6 {
        os.Exit(1)
    }
    return 0
}
