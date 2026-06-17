package main

import "io"

type Logger = struct {
    name: str
}

fn (self: #Logger) drop() {
    io.PrintLn("Dropping Logger: ${self.name}")
}

fn fail_if_odd(n: i32, log: @Logger) Result[i32, str] {
    // log is moved here, it will be dropped when this function returns
    if (n % 2 != 0) {
        return Err[i32, str]("Odd number")
    }
    return Ok[i32, str](n)
}

fn test_raii(n: i32) Result[i32, str] {
    var l1 = Logger {
        name: "Managed1"
    }
    var l2 = Logger {
        name: "Managed2"
    }

    // Move l1 into fail_if_odd. l2 remains in this scope.
    var val = fail_if_odd(n, @l1)?

    // If n is odd, fail_if_odd returns Err.
    // '?' should trigger a return from test_raii.
    // l2 MUST be dropped before test_raii returns.
    return Ok[i32, str](val)
}

fn main() i32 {
    io.PrintLn("--- Test Success ---")
    test_raii(2)

    io.PrintLn("--- Test Failure ---")
    test_raii(3)

    return 0
}
