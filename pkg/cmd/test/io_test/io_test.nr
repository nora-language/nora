package main

import "io"

fn main() i32 {
    // ── Built-in println with interpolation ───────────────────────────────────
    io.PrintLn("=== IO Standard Library Test ===")

    var greeting = "Hello, Nora!"
    io.PrintLn("String:  ${greeting}")

    var n: i32 = 42
    io.PrintLn("Int:     ${n}")

    var f: f64 = 3.14
    io.PrintLn("Float:   ${f}")

    var bt = true
    var bf = false
    io.PrintLn("Bool T:  ${bt}")
    io.PrintLn("Bool F:  ${bf}")

    // ── io.format_* → turn values into strings ────────────────────────────────
    io.PrintLn("")
    io.PrintLn("=== Format functions ===")
    var si = io.FormatInt(12345)
    io.PrintLn("format_int(12345):   ${si}")

    var sf = io.FormatFloat(2.718)
    io.PrintLn("format_float(2.718): ${sf}")

    var sb = io.FormatBool(true)
    io.PrintLn("format_bool(true):   ${sb}")

    // ── eprintln goes to stderr ───────────────────────────────────────────────
    io.EprintLn("[stderr] This goes to stderr")

    // ── File I/O ──────────────────────────────────────────────────────────────
    io.PrintLn("")
    io.PrintLn("=== File I/O ===")

    var path = "pkg/cmd/test/io_test_tmp.txt"
    var open_write_res = io.OpenFile(path, "w")
    match open_write_res {
        Ok(fw) => {
            fw.Write("line one\n")
            fw.Write("line two\n")
            fw.Write("line three\n")
            io.PrintLn("Wrote 3 lines to ${path}")
        }
        Err(err) => {
            io.PrintLn("failed to open file for writing")
        }
    }

    var open_read_res = io.OpenFile(path, "r")
    match open_read_res {
        Ok(fr) => {
            var line_opt = fr.ReadLine()
            while line_opt.IsSome() {
                var line = line_opt.Unwrap()
                io.Print("Read: " + line)
                line_opt = fr.ReadLine()
            }
        }
        Err(err) => {
            io.PrintLn("failed to open file for reading")
        }
    }

    io.PrintLn("All IO tests completed.")
    return 0
}
