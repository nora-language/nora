package main

import "io"

fn worker(id: i32, task: i32, ch: #chan[i32]) {
    io.PrintLn("Worker ${id} starting task ${task}")
    ch <- id * task
}

fn main() i32 {
    var ch = make(chan[i32], 10)

    scope {

        spawn worker(1, 100, #ch)
        spawn worker(2, 200, #ch)

        var r1 = <-ch
        var r2 = <-ch

        io.PrintLn("Results: ${r1}, ${r2}")

        return 0


    }
}
