package main

import "io"
import "sync"

fn worker(id: i32, wg: #sync.WaitGroup) {
    var i = 0
    while (i < 50) {
        i = i + 1
    }
    wg.Done()
}

fn main() i32 {
    io.PrintLn("--- Starting Scheduler Concurrency Cleanup Test ---")
    
    var wg = sync.NewWaitGroup()
    wg.Add(20)
    
    var i = 0
    while (i < 20) {
        spawn worker(i, #wg)
        i = i + 1
    }
    
    wg.Wait()
    io.PrintLn("Workers finished successfully.")
    return 0;
}
