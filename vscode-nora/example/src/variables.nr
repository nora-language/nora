package main

import "io"

fn variables() {
    var a = "initial"
    io.PrintLn(a)

    var b = 1
    var c = 3
    io.PrintLn(b, c)

    var d = true
    io.PrintLn(d)

    var e: i32 = 0
    io.PrintLn(e)

    var g = 0.32
    io.PrintLn("the number is ${g}")
    var h = 0.32
    io.PrintLn("the number is ${h}")
    var i = 0.32
    io.PrintLn("the number is ${i}")
}

fn cobaPar(a: i32, b: str) {
    io.PrintLn("${a} ${b}")
}
