package main

import "io"
import "sync"

[shared]
pub type PublicWorker = interface {
    fn Work() void
}

[shared]
type MyWorker = struct {
    id: i32
}

fn (_w: #MyWorker) Work() void {
    io.PrintLn("PublicWorker working...")
}

fn run_worker(w: #PublicWorker, wg: #sync.WaitGroup) {
    w.Work()
    wg.Done()
}

fn main() i32 {
    io.PrintLn("--- Repro Public Interface Test ---")
    var wk: PublicWorker = MyWorker { id: 42 }
    var wg = sync.NewWaitGroup()
    wg.Add(1)
    
    // Pass #wk (read lease of PublicWorker interface) across spawn boundary
    scope {
        spawn run_worker(#wk, #wg)

        wg.Wait()
        io.PrintLn("Success!")
        return 0

    }
}
