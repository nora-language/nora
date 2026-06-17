package main

import "io"

fn pinger(ch: #chan[i32]) {
    io.PrintLn("Pinger: Sending ping...", 0)
    ch <- 42
    io.PrintLn("Pinger: Ping delivered!", 0)
}

fn main() i32 {
    // Unbuffered channel (capacity 0)
    var ch = make(chan[i32], 0)

    scope {

        spawn pinger(#ch)

        io.PrintLn("Main: Waiting for ping...", 0)
        var val = <-ch
        io.PrintLn("Main: Received ping: ${val}")

        return 0


    }
}
