package main

import "io"

fn main() {
    var b = "goc" + "c"

    io.PrintLn(b)
    io.PrintLn("goc" + "dd")
    io.PrintLn("goc", "dd")

    io.PrintLn("1+1 =", 1 + 1)
    io.PrintLn("7.0/3.0 =", 7.0 / 3.0)

    io.PrintLn("Concatenate str and bool:")
    io.PrintLn("True: " + true)
    io.PrintLn("False: " + false)
    io.PrintLn(true + " is truthy")

    io.PrintLn(true && false)
    io.PrintLn(true || false)
    io.PrintLn(!true)
    io.PrintLn(true)
}
