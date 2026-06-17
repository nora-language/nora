package main

import "io"

fn worker(done: #chan[bool]) {
    io.PrintLn("Worker: Starting...")
    park()
    io.PrintLn("Worker: Resumed!")
    park()
    io.PrintLn("Worker: Finishing...")
    done <- true
}

fn main() {
    scope {
        io.PrintLn("Main: Spawning worker...")
        var done = make(chan[bool], 1)
        var f = spawn worker(#done)

        io.PrintLn("Main: Resuming worker (1st time)...")
        resume(f)

        io.PrintLn("Main: Back in main, resuming worker (2nd time)...")
        resume(f)

        io.PrintLn("Main: Back in main, resuming worker (3rd time)...")
        resume(f)

        io.PrintLn("Main: Waiting for worker to finish...")
        <-done
        io.PrintLn("Main: Done.")
    }
}
