package main

import "io"

fn sender1(ch: #chan[i32]) {
    ch <- 100
}

fn sender2(ch: #chan[i32]) {
    ch <- 200
}

fn main() i32 {
    var c1 = make(chan[i32])
    var c2 = make(chan[i32])

    scope {
        spawn sender1(#c1)
        spawn sender2(#c2)

        var count = 0
        var sum = 0
        while count < 2 {
            var val = 0
            select {
                case val = <-c1:
                    sum = sum + val
                case val = <-c2:
                    sum = sum + val
            }
            count++
        }

        io.PrintLn("Fiber channel select test passed: sum=", sum)
        return 0
    }
}
