package main

import "io"

extern fn get_worker_id() i32

fn task_success() Result[bool, str] {
    io.PrintLn("Task Success: Executing on Worker ${get_worker_id()}")
    return Ok[bool, str](true)
}

fn task_failure() Result[bool, str] {
    io.PrintLn("Task Failure: Executing on Worker ${get_worker_id()}")
    return Err[bool, str]("API Connection Failed")
}

fn main() i32 {
    io.PrintLn("=== Nora Structured Concurrency: Parallel Error Propagation ===")

    var res = parallel {
        task_success()
        task_failure()
    }

    match res {
        Ok(_) => {
            panic("Error propagation failed: parallel block should return Err")
        }
        Err(err) => {
            io.PrintLn("Successfully propagated parallel error: " + err)
        }
    }

    return 0
}
