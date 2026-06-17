package main

import "io"

/// stupid coba dong
fn loop() {
    var max: i32 = 0
    while (max <= 10) {
        io.PrintLn("value : ${max}")
        max++
    }
}
