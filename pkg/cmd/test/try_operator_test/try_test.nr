package main

import "io"

type Resource = struct {
    id: i32,
    name: str
}

fn (self: #Resource) drop() {
    io.PrintLn("Resource drop: ${self.name} (id: ${self.id})")
}

// Result Try Tests
fn success_result(x: i32) Result[i32, str] {
    return Ok[i32, str](x)
}

fn fail_result(x: i32) Result[i32, str] {
    return Err[i32, str]("Result failure code: ${x}")
}

fn test_result_success() Result[i32, str] {
    var a = success_result(10)?
    var b = success_result(20)?
    return Ok[i32, str](a + b)
}

fn test_result_fail() Result[i32, str] {
    var a = success_result(10)?
    var b = fail_result(5)?
    // This line should not be reached
    return Ok[i32, str](a + b)
}

// Option Try Tests
fn success_option(x: i32) Option[i32] {
    return Some[i32](x)
}

fn fail_option() Option[i32] {
    return None[i32]
}

fn test_option_success() Option[i32] {
    var a = success_option(100)?
    var b = success_option(200)?
    return Some[i32](a + b)
}

fn test_option_fail() Option[i32] {
    var a = success_option(100)?
    var b = fail_option()?
    // This line should not be reached
    return Some[i32](a + b)
}

// Inline complex expressions with try operator
fn test_complex_result() Result[i32, str] {
    var sum = success_result(5)? + success_result(15)?
    return Ok[i32, str](sum)
}

fn test_complex_option() Option[i32] {
    var sum = success_option(7)? + success_option(8)?
    return Some[i32](sum)
}

// RAII Cleanups on Early Return
fn process_with_raii_result(should_fail: bool) Result[i32, str] {
    var _res1 = Resource {
        id: 1,
        name: "LocalRes1"
    }
    var _res2 = Resource {
        id: 2,
        name: "LocalRes2"
    }

    if (should_fail) {
        // Move res1 to a function that succeeds
        var _val1 = success_result(42)?

        // This fails and returns early. res2 MUST be dropped here!
        var _val2 = fail_result(99)?
    }

    return Ok[i32, str](100)
}

fn process_with_raii_option(should_fail: bool) Option[i32] {
    var _res1 = Resource {
        id: 3,
        name: "LocalRes3"
    }
    var _res2 = Resource {
        id: 4,
        name: "LocalRes4"
    }

    if (should_fail) {
        // Move res1 to a function that succeeds
        var _val1 = success_option(42)?

        // This fails and returns early. res2 MUST be dropped here!
        var _val2 = fail_option()?
    }

    return Some[i32](100)
}

fn main() i32 {
    io.PrintLn("=== Starting try operator tests ===")

    // Result basic tests
    var r1 = test_result_success()
    var r1_ok = r1.IsOk()
    var r1_val = r1.Unwrap()
    io.PrintLn("Result Success IsOk: ${r1_ok}, Value: ${r1_val}")

    var r2 = test_result_fail()
    io.PrintLn("Result Fail IsErr: ${r2.IsErr()}")
    match r2 {
        Ok(val) => {
            io.PrintLn("Unexpected Success: ${val}")
        }
        Err(msg) => {
            io.PrintLn("Result Err: ${msg}")
        }
        _ => {
            io.PrintLn("Unknown Result state")
        }
    }

    // Option basic tests
    var o1 = test_option_success()
    var o1_some = o1.IsSome()
    var o1_val = o1.Unwrap()
    io.PrintLn("Option Success IsSome: ${o1_some}, Value: ${o1_val}")

    var o2 = test_option_fail()
    io.PrintLn("Option Fail IsNone: ${o2.IsNone()}")

    // Complex expressions
    var cr = test_complex_result()
    io.PrintLn("Complex Result: ${cr.Unwrap()}")

    var co = test_complex_option()
    io.PrintLn("Complex Option: ${co.Unwrap()}")

    // RAII Drop tests
    io.PrintLn("--- RAII Result Success Flow ---")
    var raii_r1 = process_with_raii_result(false)
    io.PrintLn("RAII Result Success IsOk: ${raii_r1.IsOk()}")

    io.PrintLn("--- RAII Result Failure Flow ---")
    var raii_r2 = process_with_raii_result(true)
    io.PrintLn("RAII Result Failure IsErr: ${raii_r2.IsErr()}")

    io.PrintLn("--- RAII Option Success Flow ---")
    var raii_o1 = process_with_raii_option(false)
    io.PrintLn("RAII Option Success IsSome: ${raii_o1.IsSome()}")

    io.PrintLn("--- RAII Option Failure Flow ---")
    var raii_o2 = process_with_raii_option(true)
    io.PrintLn("RAII Option Failure IsNone: ${raii_o2.IsNone()}")

    io.PrintLn("=== All tests finished ===")
    return 0
}
