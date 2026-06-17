package main

import "io"

fn main() i32 {
    var name = "Nora"
    var version: i32 = 2
    var score: i32 = 99
    var pi = 3.14159
    var passed = true

    // ── Interpolation inside io.PrintLn(zero-alloc: direct printf) ──────────────
    io.PrintLn("=== String Interpolation ===")
    io.PrintLn("Language: ${name} v${version}")
    io.PrintLn("Score: ${score}, Passed: ${passed}")
    io.PrintLn("Pi is approximately ${pi}")

    var a: i32 = 5
    var b: i32 = 10
    io.PrintLn("${a} + ${b} = ${a + b}")
    io.PrintLn("${a} * ${b} = ${a * b}")

    // ── Mixed interpolation and values ────────────────────────────────────────
    io.PrintLn("=== Mixed ===")
    io.PrintLn("${name} scored ${score} out of 100")
    io.PrintLn("Flags: ${passed}")

    // ── Multi-arg (Go-style space-separated) ─────────────────────────────────
    io.PrintLn("=== Multi-arg ===")
    io.PrintLn(name, "version", version)
    io.PrintLn("result:", a + b, "passed:", passed)

    // ── Plain strings ─────────────────────────────────────────────────────────
    io.PrintLn("=== Types ===")
    io.PrintLn(42)
    io.PrintLn(3.14)
    io.PrintLn(true)
    io.PrintLn("just a string")

    // ── eprintln with interpolation ───────────────────────────────────────────
    io.EprintLn("[stderr] ${name} v${version} error check")

    io.PrintLn("Done.")
    return 0
}
