package main

import "io"

extern fn get_worker_id() i32

extern fn Sleep(dwMilliseconds: i32)

fn fibonacci(n: i32) i32 {
    if (n <= 1) {
        return n
    }
    return fibonacci(n - 1) + fibonacci(n - 2)
}

fn worker_a(id: i32, ch: #chan[i32]) {
    io.PrintLn("Worker A${id} starting on Scheduler Worker ${get_worker_id()}")
    var sum = 0
    var i = 0
    while i < 50 {
        sum = sum + fibonacci(i % 10 + 10)
        i++
    }
    ch <- sum
    io.PrintLn("Worker A${id} completed on Scheduler Worker ${get_worker_id()}")
}

fn task_a() {
    io.PrintLn("Task A starting on Scheduler Worker ${get_worker_id()}")
    var ch1 = make(chan[i32], 1)
    var ch2 = make(chan[i32], 1)

    scope {

        spawn worker_a(1, #ch1)
        spawn worker_a(2, #ch2)

        var val1 = <-ch1
        var val2 = <-ch2
        var total = val1 + val2

        io.PrintLn("Task A completed! Computed sub-total: ${total}")


    }
}

fn worker_b(id: i32, ch: #chan[i32]) {
    io.PrintLn("Worker B${id} starting on Scheduler Worker ${get_worker_id()}")
    var sum = 0
    var i = 0
    while i < 50 {
        sum = sum + fibonacci(i % 10 + 12) // Slightly heavier computation
        i++
    }
    ch <- sum
    io.PrintLn("Worker B${id} completed on Scheduler Worker ${get_worker_id()}")
}

fn task_b() {
    io.PrintLn("Task B starting on Scheduler Worker ${get_worker_id()}")
    var ch1 = make(chan[i32], 1)
    var ch2 = make(chan[i32], 1)

    scope {
        spawn worker_b(1, #ch1)
        spawn worker_b(2, #ch2)

        var val1 = <-ch1
        var val2 = <-ch2
        var total = val1 + val2

        if (total != 60) {
            io.PrintLn("ERROR: task_b failed. Expected 60, got ${total}")
        }
    }
}

fn main() i32 {
    io.PrintLn("=== Nora High-Performance Concurrency & Parallelism Demo ===")
    io.PrintLn("Main starting on Scheduler Worker ${get_worker_id()}")

    // These run physically in parallel on separate OS thread workers!
    parallel {
        task_a()
        task_b()
    }

    io.PrintLn("Main: Both physical parallel tasks completed successfully!")
    io.PrintLn("Main finishing on Scheduler Worker ${get_worker_id()}")
    return 0
}
