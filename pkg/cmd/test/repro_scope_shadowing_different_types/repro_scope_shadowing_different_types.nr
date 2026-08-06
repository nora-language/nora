package main

import "io"
import "os"

fn main() i32 {
    var x: i32 = 42

    scope {
        // Inner x shadows outer x with different type (str)
        var x: str = "shadowed_string"
        if x != "shadowed_string" {
            io.PrintLn("Error: Inner shadowed string mismatch")
            os.Exit(1)
        }
    }

    // Outer x must retain its original type (i32) and value (42)
    if x != 42 {
        io.PrintLn("Error: Outer variable corrupted by scope shadowing")
        os.Exit(1)
    }

    io.PrintLn("Scope shadowing test passed: outer x=", x)
    return 0
}
