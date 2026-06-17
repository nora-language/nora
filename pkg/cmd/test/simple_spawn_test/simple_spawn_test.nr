package main

import "io"

fn sender(c: #chan[i32]) {
    io.PrintLn("Fiber 1: sending 42")
    c <- 42
}

fn receiver(c: #chan[i32], done: #chan[bool]) {
    var val = <-c
    io.PrintLn("Fiber 2: received ${val}")
    done <- true
}

fn main() {
    io.PrintLn("Hello from Nora Wasm!")

    var c = make(chan[i32], 0)
    var done = make(chan[bool], 0)

    scope {

        spawn sender(#c)
        spawn receiver(#c, #done)

        var a = <-done
        io.PrintLn("Main fiber done ${a}")


    }
}
