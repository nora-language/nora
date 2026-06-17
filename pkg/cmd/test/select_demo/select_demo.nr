package main

import "io"

fn producer(id: i32, ch: #chan[i32]) {
    var i = 0
    while i < 5 {
        ch <- id * 10 + i
        i++
    }
}

fn main() i32 {
    var c1 = make(chan[i32])
    var c2 = make(chan[i32])

    scope {

        spawn producer(1, #c1)
        spawn producer(2, #c2)

        var v = 0
        var i = 0
        while i < 10 {
            select {
                case v = <-c1:
                    io.PrintLn("From C1: ${v}")
                case v = <-c2:
                    io.PrintLn("From C2: ${v}")
            }
            i++
        }

        return 0


    }
}

