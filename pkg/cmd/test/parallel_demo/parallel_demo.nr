package main

import "io"

extern fn get_worker_id() i32

fn worker1(done: #chan[bool]) {
    var wid = get_worker_id()
    io.PrintLn("Worker 1 running on Scheduler Worker ${wid}")
    park()
    io.PrintLn("Worker 1 resumed on Scheduler Worker ${get_worker_id()}")
    park()
    done <- true
}

fn worker2(done: #chan[bool]) {
    var wid = get_worker_id()
    io.PrintLn("Worker 2 running on Scheduler Worker ${wid}")
    park()
    io.PrintLn("Worker 2 resumed on Scheduler Worker ${get_worker_id()}")
    park()
    done <- true
}

fn main() {
    var wid = get_worker_id()
    io.PrintLn("Main starting on Scheduler Worker ${wid}")

    var d1 = make(chan[bool], 1)
    var d2 = make(chan[bool], 1)

    scope {
        var f1 = spawn worker1(#d1)
        var f2 = spawn worker2(#d2)

        resume(f1)
        resume(f2)

        resume(f1)
        resume(f2)

        <-d1
        <-d2
    }

    io.PrintLn("Main finishing on Scheduler Worker ${get_worker_id()}")
}
