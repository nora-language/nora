package main

import "io"

fn worker(id: i32, jobs: #chan[i32], results: #chan[i32]) {
    for  {
        var job = <-jobs // Parks if no jobs
        if (job == -1) {
            break
        }
        io.PrintLn("Worker ${id} processing ${job}")
        results <- job * 2 // Sends result back
    }
}

fn main() {
    var jobs = make(chan[i32], 100)
    var results = make(chan[i32], 100)

    // Wrap the entire lifecycle in scope so we can pass leases safely
    scope {
        // Spawn 4 parallel workers (M:N at work)
        var i = 0
        while i < 4 {
            spawn worker(i, #jobs, #results)
            i++
        }

        i = 0
        while i < 100 {
            jobs <- 10
            i++
        }
        
        // Wait for all 100 results before stopping
        i = 0
        while i < 100 {
            var res = <-results
            i++
        }
        io.PrintLn("All 100 jobs processed successfully!")

        // Stop workers
        i = 0
        while i < 4 {
            jobs <- -1
            i++
        }
    }
}

