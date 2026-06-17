package main

import "io"

fn worker(x: #i32) {
    io.PrintLn(x.ToString())
}

fn main() i32 {
    var x = 10
    
    scope {
        // Violation: passing non-exempt lease inside scope block
        // (Even in scope block, only [shared] structs or channels can cross)
        scope {
            spawn worker(#x)

        }
}
    
    return 0
}
