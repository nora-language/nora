// Gap 4.2d: strconv.ParseInt (std/strconv/strconv.nr:89) checks positive overflow against 2147483647
// before applying the negative sign, rejecting valid INT32_MIN ("-2147483648").
package main

import "strconv"
import "os"

fn main() i32 {
    var s = "-2147483648"
    match strconv.ParseInt(#s, 10) {
        Ok(v) => {
            if v != -2147483647 - 1 {
                os.Exit(1)
            }
        }
        Err(e) => {
            os.Exit(1)
        }
    }
    return 0
}
