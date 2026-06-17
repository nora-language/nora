package main

import "io"
import "time"

fn main() {
    io.PrintLn("=== Starting Nora Ticker Premium Stress Test ===")

    // Create a 50ms ticker (stack duration is 100% leak-free!)
    var ms_dur = time.Millisecond()
    var ticker_dur = time.Duration {
        nanoseconds: ms_dur.Milliseconds() * 50 * 1000000
    }
    var ticker = time.NewTicker(ticker_dur)

    io.PrintLn("Created ticker with 50ms interval. Collecting 3 ticks...")

    var t_start = time.Now()

    // Declare loop variable outside the loop to ensure clean drop during reassignment!
    var tick_time = time.Now()

    // Receive 3 ticks
    var i = 0
    while i < 3 {
        tick_time = <-ticker.C
        var elapsed = (tick_time.UnixNano() - t_start.UnixNano()) / 1000000
        io.PrintLn("  Tick ${i} received at epoch ${tick_time.Unix()}. Elapsed: ${elapsed}ms")
        i++
    }

    io.PrintLn("Stopping the ticker...")
    ticker.Stop()

    // Sleep for 150ms to ensure no ticks are queued or sent anymore
    var sleep_dur = time.Duration {
        nanoseconds: ms_dur.Milliseconds() * 150 * 1000000
    }
    time.Sleep(sleep_dur)

    // Drain and verify that channel has no more events (non-blocking receive)
    var unexpected_ticks = 0
    select {
        case tick_time = <-ticker.C:
            unexpected_ticks = unexpected_ticks + 1
        default:
        // No ticks, this is exactly what we want!
    }

    if (unexpected_ticks > 0) {
        panic("Error: received tick after Stop() was called")
    }

    io.PrintLn("Nora Ticker Premium Stress Test passed successfully!")
}
