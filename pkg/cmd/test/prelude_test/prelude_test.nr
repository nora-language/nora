package main

import "io"

fn test_option() {
    var o1 = Some[i32](42)
    var o2 = None[i32]

    io.PrintLn("o1.IsSome(): ${o1.IsSome()}")
    io.PrintLn("o1.IsNone(): ${o1.IsNone()}")
    io.PrintLn("o2.IsSome(): ${o2.IsSome()}")
    io.PrintLn("o2.IsNone(): ${o2.IsNone()}")

    var v1 = o1.Unwrap()
    io.PrintLn("o1.Unwrap(): ${v1}")

    var v2 = o2.UnwrapOr(100)
    io.PrintLn("o2.UnwrapOr(): ${v2}")
}

fn test_result() {
    var r1 = Ok[i32, str](10)
    var r2 = Err[i32, str]("something went wrong")

    io.PrintLn("r1.IsOk(): ${r1.IsOk()}")
    io.PrintLn("r1.IsErr(): ${r1.IsErr()}")
    io.PrintLn("r2.IsOk(): ${r2.IsOk()}")
    io.PrintLn("r2.IsErr(): ${r2.IsErr()}")

    var v1 = r1.Unwrap()
    io.PrintLn("r1.Unwrap(): ${v1}")

    var v2 = r2.UnwrapOr(999)
    io.PrintLn("r2.UnwrapOr(): ${v2}")
}

fn divide(a: i32, b: i32) Result[i32, str] {
    if (b == 0) {
        return Err[i32, str]("division by zero")
    }
    return Ok[i32, str](a / b)
}

fn compute_val(a: i32, b: i32) Result[i32, str] {
    var val = divide(a, b)?
    return Ok[i32, str](val * 2)
}

fn test_try_operator() {
    var r1 = compute_val(10, 2)
    var r2 = compute_val(10, 0)

    if (r1.IsOk()) {
        io.PrintLn("r1 succeeded: ${r1.Unwrap()}")
    } else {
        io.PrintLn("r1 failed: ${r1.UnwrapOr(-1)}")
    }

    if (r2.IsOk()) {
        io.PrintLn("r2 succeeded: ${r2.Unwrap()}")
    } else {
        io.PrintLn("r2 failed: ${r2.UnwrapOr(-1)}")
    }
}

fn main() i32 {
    io.PrintLn("=== Testing Standard Option/Result Prelude ===")
    test_option()
    test_result()
    test_try_operator()
    io.PrintLn("=== Done ===")
    return 0
}
