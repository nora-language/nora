package main

import "io"

fn main() {
    var name = "Nora"
    var version = 1

    var greeting = "Hello, ${name}! Version: ${version}"
    io.PrintLn(greeting)

    var a = 5
    var b = 10
    var calc = "Result: ${a + b} is greater than ${a}"
    io.PrintLn(calc)

    var flag = true
    var status = "Status is ${flag}"
    io.PrintLn(status)

    // not working
    io.PrintLn("Result: ${a + b} is greater than ${a}")
}
