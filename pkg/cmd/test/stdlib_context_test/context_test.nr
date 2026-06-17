package main

import "io"
import "context"
import "time"
import "strconv"

fn sleep_ms(ms: i32) {
    var ms_dur = time.Millisecond()
    var i = 0
    while (i < ms) {
        time.Sleep(#ms_dur)
        i = i + 1
    }
}

fn background_worker(id: i32, ctx: #context.Context) {
    while (true) {
        if (ctx.IsCancelled()) {
            io.PrintLn("Worker " + strconv.Itoa(id) + " gracefully stopping due to: " + ctx.Err())
            return
        }
        sleep_ms(50)
    }
}

fn main() {
    var ctx = context.WithCancel()

    io.PrintLn("Starting workers...")
    spawn background_worker(1, #ctx)
    spawn background_worker(2, #ctx)
    spawn background_worker(3, #ctx)

    sleep_ms(200)

    io.PrintLn("Cancelling context...")
    (&ctx).Cancel("Shutdown requested")

    // Give workers time to cleanly exit
    sleep_ms(100)
    io.PrintLn("Main thread finished successfully.")
}
