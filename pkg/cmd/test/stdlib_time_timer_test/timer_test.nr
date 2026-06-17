package main

import "time"
import "io"

fn main() i32 {
    io.PrintLn("=== Testing time.After ===")
    var d50 = time.Duration { nanoseconds: 50 * 1000000 }
    var c = time.After(d50)
    
    select {
        case <-c:
            io.PrintLn("time.After fired successfully.")
        case <-time.After(time.Duration { nanoseconds: 100 * 1000000 }):
            panic("Failed: 100ms fired before 50ms!")
    }

    io.PrintLn("=== Testing time.Timer Stop ===")
    var timer = time.NewTimer(time.Duration { nanoseconds: 100 * 1000000 })
    timer.Stop()
    
    var c2 = time.After(time.Duration { nanoseconds: 50 * 1000000 })
    
    select {
        case <-timer.C:
            panic("Failed: Stopped timer fired!")
        case <-c2:
            io.PrintLn("Stopped timer did not fire.")
    }

    io.PrintLn("All timer tests passed.")
    return 0
}
