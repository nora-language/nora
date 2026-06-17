package main

import "io"
import "nursery"

extern fn get_worker_id() i32

fn success_worker(n: #nursery.Context, id: i32) {
    io.PrintLn("Dynamic worker ${id} starting on scheduler worker ${get_worker_id()}...")
    n.done("")
}

fn failure_worker(n: #nursery.Context, id: i32) {
    io.PrintLn("Dynamic worker ${id} starting on scheduler worker ${get_worker_id()}...")
    n.done("Connection timed out on task ${id}")
}

fn panic_worker(n: #nursery.Context, id: i32) {
    io.PrintLn("Dynamic worker ${id} starting on scheduler worker ${get_worker_id()}...")
    panic("Unexpected fatal error in worker ${id}")
}

fn main() i32 {
    io.PrintLn("=== Nora Dynamic nursery: Structured Concurrency ===")

    var n = nursery.New()

    // Spawn 3 successful workers
    var i = 0
    while i < 3 {
        n.start_task()
        spawn(n.completion_chan) success_worker(#n, i)
        i++
    }

    // Spawn 1 panicking worker
    n.start_task()
    spawn(n.completion_chan) panic_worker(#n, 99)

    var res = n.wait()

    match res {
        Ok(_) => {
            panic("nursery wait failed: should return Err due to panic_worker")
        }
        Err(err) => {
            io.PrintLn("Successfully aggregated dynamic nursery error: " + err)
        }
    }

    return 0
}
