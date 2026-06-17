package main

import "io"

fn generate(ch: #chan[i32]) {
    var i = 2
    while i < 1000 {
        ch <- i
        i++
    }
}

fn filter(in_ch: #chan[i32], out_ch: #chan[i32], prime: i32) {
    for  {
        var i = <-in_ch
        if (i % prime != 0) {
            out_ch <- i
        }
    }
}

fn main() i32 {
    io.PrintLn("--- Concurrent Prime Sieve (50 Primes) ---")
    var ch = make(chan[i32], 100)
    scope {
        spawn generate(#ch)

        var i = 0
        while i < 50 {
            var prime = <-ch
            io.PrintLn("Prime ${i + 1}: ${prime}")

            var next_ch = make(chan[i32], 100)
            spawn filter(#ch, #next_ch, prime)
            ch = next_ch
            i++
        }

        return 0

    }
}
