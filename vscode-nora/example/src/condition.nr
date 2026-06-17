package main

import "io"

fn condition() {
    var a = 20
    var b = 20
    if (a == b) {
        io.PrintLn("angka 20")
    } else {
        io.PrintLn("angka 300")
    }
}
