package main

import "io"

fn producer(ch: #chan[i32], count: i32) {
    {
        var i = 0
        while i < count {
        ch <- i
            i++
        }
    }
}

fn main() i32 {
    var P = 5000
    var C = 20
    io.PrintLn("--- Concurrency Benchmark: Flood ---")
    io.PrintLn("Producers: ${P}")
    io.PrintLn("Messages per producer: ${C}")

    var ch = make(chan[i32], 1000)
    scope {
        {
            var i = 0
            while i < P {
            spawn producer(#ch, C)
                i++
            }
        }

        var total = P * C
        {
            var j = 0
            while j < total {
            var val = <-ch
                j++
            }
        }
    }

    io.PrintLn("Flood benchmark finished!")
    return 0
}

