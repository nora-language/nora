package main

import "io"
import "time"

extern fn get_worker_id() i32

fn starvation_worker(ch: #chan[bool], exit_ch: #chan[bool]) {
    io.PrintLn("Starvation worker starting on Scheduler Worker ${get_worker_id()}")
    var counter = 0
    while (true) {
        counter = counter + 1
        select {
            case <-ch:
                io.PrintLn("Starvation worker received stop signal after ${counter} loops! Exiting.")
                exit_ch <- true
                return
            default:
            // Continue looping
        }
    }
}

fn check_worker(done_ch: #chan[bool]) {
    io.PrintLn("Check worker starting on Scheduler Worker ${get_worker_id()}")
    var d = time.Millisecond()
    time.Sleep(d)
    done_ch <- true
    io.PrintLn("Check worker completed on Scheduler Worker ${get_worker_id()}")
}

fn main() i32 {
    io.PrintLn("=== Nora Anti-Starvation Cooperative Preemption Test ===")

    var stop_ch = make(chan[bool], 1)
    var done_ch = make(chan[bool], 1)
    var exit_ch = make(chan[bool], 1)

    scope {

        spawn starvation_worker(#stop_ch, #exit_ch)
        spawn check_worker(#done_ch)

        var ok = <-done_ch
        if (ok) {
            io.PrintLn("Main: Check worker ran successfully! Stopping starvation worker.")
            stop_ch <- true

            // Wait for starvation worker to exit cleanly
            <-exit_ch
        }

        io.PrintLn("Main completed successfully!")
        return 0


    }
}
