package main

import "io"
import "nursery"
import "os"

fn sync_worker(n: #nursery.Context, id: i32, ch: #chan[i32]) {
    ch <- id * 10
    n.done("")
}

fn main() i32 {
    var n = nursery.New()
    var ch = make(chan[i32])

    n.start_task()
    spawn(n.completion_chan) sync_worker(#n, 1, #ch)

    n.start_task()
    spawn(n.completion_chan) sync_worker(#n, 2, #ch)

    var v1 = <-ch
    var v2 = <-ch

    var res = n.wait()
    match res {
        Ok(_) => {
            if (v1 + v2) != 30 {
                io.PrintLn("Error: Nursery worker channel sum wrong")
                os.Exit(1)
            }
            io.PrintLn("Nursery channel sync test passed: sum=", v1 + v2)
        }
        Err(e) => {
            io.PrintLn("Error: Nursery wait failed: ${e}")
            os.Exit(1)
        }
    }
    return 0
}
