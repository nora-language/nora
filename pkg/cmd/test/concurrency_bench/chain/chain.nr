package main

import "io"

fn relay(in_ch: #chan[i32], out_ch: #chan[i32], rounds: i32) {
    {
        var i = 0
        while i < rounds {
        var val = <-in_ch
        out_ch <- val
            i++
        }
    }
}

fn main() i32 {
    var N = 1000
    var M = 500
    io.PrintLn("--- Concurrency Benchmark: Chain ---")
    io.PrintLn("Fibers: ${N}")
    io.PrintLn("Tokens: ${M}")

    var start_ch = make(chan[i32], 1)

    var c1 = make(chan[i32], 1)
    scope {
        spawn relay(#start_ch, #c1, M)
        var prev_ch = c1

        {
            var i = 1
            while i < N {
            var next_ch = make(chan[i32], 1)
            spawn relay(#prev_ch, #next_ch, M)
            prev_ch = next_ch
                i++
            }
        }

        {
            var j = 0
            while j < M {
            start_ch <- j
            var res = <-prev_ch
                j++
            }
        }

        io.PrintLn("Chain benchmark finished!")
        return 0

    }
}

