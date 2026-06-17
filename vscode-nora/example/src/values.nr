package main

import "io"

fn values() {
    var b = true
    var a = "stupd" + "helo ${b}"
    io.PrintLn("${b} ${a}")
    io.PrintLn("goc", "dd")

    io.PrintLn("1+1 =", 1 + 1)
    io.PrintLn("7.0/3.0 =", 7.0 / 3.0)

    io.PrintLn(true && false)
    io.PrintLn(true || false)
    io.PrintLn(!true)
}
