package main

import "io"
import "time"

fn main() {
    io.PrintLn("--- Starting std/time Premium Test Suite ---")

    // 1. Check Precision Now
    var t1 = time.Now()
    var u_sec = t1.Unix()
    var u_nsec = t1.UnixNano()
    io.PrintLn("Current Epoch Secs: ${u_sec}")
    io.PrintLn("Current Epoch Nsecs: ${u_nsec}")

    // 2. Formatting Check
    var date_str = t1.Format("%Y-%m-%d")
    io.PrintLn("Formatted Date: ${date_str}")

    // 3. Calendar Breakdown
    var local = t1.Local()
    var yr = local.year
    var mo = local.month
    var dy = local.day
    var hr = local.hour
    var mn = local.minute
    var sc = local.second
    io.PrintLn("Local Breakdown:")
    io.PrintLn("  Year:   ${yr}")
    io.PrintLn("  Month:  ${mo}")
    io.PrintLn("  Day:    ${dy}")
    io.PrintLn("  Time:   ${hr}:${mn}:${sc}")

    // 4. Durations and High-Precision Sleep
    io.PrintLn("Sleeping for 150ms...")
    var ms_dur = time.Millisecond()
    var sleep_dur = ms_dur.Milliseconds() * 150
    var dur = time.Duration {
        nanoseconds: sleep_dur * 1000000
    }
    time.Sleep(dur)

    var t2 = time.Now()
    var diff = t2.Sub(t1)
    var diff_ms = diff.Milliseconds()
    io.PrintLn("Elapsed Milliseconds: ${diff_ms}")

    if (diff_ms < 140) {
        panic("Error: Sleep elapsed time too short")
    }

    // 5. Time Arithmetic
    var hour = time.Hour()
    var t3 = t1.Add(hour)
    var delta = t3.Sub(t1)
    var delta_sec = delta.Seconds()
    if (delta_sec != 3600.0) {
        panic("Error: Duration delta mismatch after Add")
    }

    // 6. Time Comparisons
    if (!t1.Before(t2)) {
        panic("Error: Before() returned false")
    }
    if (!t2.After(t1)) {
        panic("Error: After() returned false")
    }
    if (!t1.Eq(t1)) {
        panic("Error: Eq() returned false for identity comparison")
    }

    io.PrintLn("std/time Premium Test Suite passed successfully!")
}
