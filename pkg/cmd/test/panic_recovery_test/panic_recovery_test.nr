package main

import "io"
import "time"

fn sleep_ms(ms: i32) {
    var ms_dur = time.Millisecond()
    var i = 0
    while (i < ms) {
        time.Sleep(#ms_dur)
        i = i + 1
    }
}

fn panic_worker() {
    sleep_ms(100)
    io.PrintLn("Worker about to panic...")
    // Trigger panic without allocating memory
    panic("manual panic")
    io.PrintLn("Unreachable")
}

fn main() {
    io.PrintLn("Main thread starting...")
    spawn panic_worker()

    // Sleep long enough for the worker to panic
    sleep_ms(500)
    
    io.PrintLn("Main thread successfully survived the worker's panic!")
}
