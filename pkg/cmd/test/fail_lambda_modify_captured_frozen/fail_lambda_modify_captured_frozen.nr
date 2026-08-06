package main

import "io"

fn main() i32 {
    var counter = 10
    
    var increment = fn() i32 {
        // ERROR: Cannot modify frozen variable 'counter' captured in closure
        counter = counter + 5
        return counter
    }

    var v1 = increment()
    io.PrintLn("v1:", v1)
    return 0
}
