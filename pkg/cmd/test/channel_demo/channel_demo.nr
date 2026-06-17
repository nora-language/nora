package main

import "io"

fn producer(ch: #chan[i32]) {
    var i = 0
    while i < 5 {
        io.PrintLn("Producing ${i}")
        ch <- i
        i = i + 1
    }
}

fn main() i32 {
    // Buffered channel (capacity 2)
    var ch = make(chan[i32], 2)

    scope {

        spawn producer(#ch)

        var j = 0
        while j < 5 {
            var val = <-ch
            io.PrintLn("Consumed ${val}")
            j = j + 1
        }

        return 0


    }
}
