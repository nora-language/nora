package main

import "io"

fn panic_worker() {
    io.PrintLn("Worker starting...")
    panic("manual crash!")
    io.PrintLn("Unreachable")
}

fn array_crash_worker() {
    io.PrintLn("Array crash worker starting...")
    panic("manual crash from array worker!")
}

fn main() i32 {
    io.PrintLn("=== Testing Spawn Panic Monitor ===")

    var c1 = make(chan[str], 1)
    spawn(c1) panic_worker()

    var err1 = <-c1
    if (err1 != "") {
        io.PrintLn("Caught panic 1: " + err1)
    }

    var c2 = make(chan[str], 1)
    spawn(c2) array_crash_worker()

    var err2 = <-c2
    if (err2 != "") {
        io.PrintLn("Caught panic 2: " + err2)
    }

    return 0
}
