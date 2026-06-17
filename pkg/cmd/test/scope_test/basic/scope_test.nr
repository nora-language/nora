package main

import "io"
import "sync"
import "strconv"

fn worker(data: #sync.AtomicInt) {
    io.PrintLn("Worker modifying data")
    data.Add(1)
}

fn main() i32 {
    io.PrintLn("Testing Scope Block")
    var data = sync.NewAtomicInt(0)

    scope {
        {
            var i = 0
            while i < 5 {
                scope {
                    spawn worker(#data)
                }
                i++
            }
        }
    } // waitgroup implicitly inserted here, blocks until all 5 tasks finish!

    io.PrintLn("All workers finished. Val: " + strconv.Itoa(data.Load()))
    return 0
}

