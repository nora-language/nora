package main

import "io"
import "sync"

[shared]
type Worker = interface {
    fn Work() void
}

[shared]
type MyWorker = struct {
    id: i32
}

fn (_w: #MyWorker) Work() void {
    io.PrintLn("Worker working...")
}

fn run_worker(w: #Worker, wg: #sync.WaitGroup) {
    w.Work()
    wg.Done()
}

fn main() i32 {
    io.PrintLn("--- Repro Spawn Interface Test ---")
    var wk: Worker = MyWorker { id: 42 }
    var wg = sync.NewWaitGroup()
    wg.Add(1)
    
    // We pass #wk (read lease of Worker interface) across spawn boundary
    scope {
        spawn run_worker(#wk, #wg)

        wg.Wait()
        io.PrintLn("Success!")
        return 0

    }
}
