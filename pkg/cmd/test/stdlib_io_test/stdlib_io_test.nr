package main

import "io"

extern fn _strlen(s: str) i32

fn test_file_static_helpers() {
    var path = "test_io_static_helper.txt"
    var content = "Hello Nora Stdlib I/O! This is a static helper test."

    // Write content
    var write_res = io.WriteFile(path, content)
    if (!write_res.IsOk()) {
        panic("failed to write file using static helper")
    }

    // Read content back
    var read_res = io.ReadFile(path)
    if (!read_res.IsOk()) {
        panic("failed to read file using static helper")
    }

    var read_str = read_res.Unwrap()
    io.PrintLn("Read back content: ${read_str}")
    if (read_str != content) {
        panic("content mismatch in static helper read/write")
    }
    io.PrintLn("test_file_static_helpers PASSED!")
}

fn test_file_raii_methods() {
    var path = "test_io_raii_methods.txt"

    // Write using File struct
    var open_write_res = io.OpenFile(path, "w")
    match open_write_res {
        Ok(f) => {
            var _w1 = f.Writeln("Line 1: RAII test")
            var _w2 = f.Writeln("Line 2: RAII test")
            // f drops/closes here when match block exits!
        }
        Err(err) => {
            panic("failed to open file for writing")
        }
    }

    // Read and verify using File struct
    var open_read_res = io.OpenFile(path, "r")
    match open_read_res {
        Ok(f) => {
            var line1_opt = f.ReadLine()
            if (!line1_opt.IsSome()) {
                panic("expected line 1 to be read")
            }
            var line1 = line1_opt.Unwrap()
            io.PrintLn("Line 1: ${line1}")
            if (line1 != "Line 1: RAII test") {
                panic("line 1 content mismatch")
            }

            var line2_opt = f.ReadLine()
            if (!line2_opt.IsSome()) {
                panic("expected line 2 to be read")
            }
            var line2 = line2_opt.Unwrap()
            io.PrintLn("Line 2: ${line2}")
            if (line2 != "Line 2: RAII test") {
                panic("line 2 content mismatch")
            }

            var line3_opt = f.ReadLine()
            if (line3_opt.IsSome()) {
                panic("expected EOF / None for line 3")
            }
            // f drops/closes here when match block exits!
        }
        Err(err) => {
            panic("failed to open file for reading")
        }
    }
    io.PrintLn("test_file_raii_methods PASSED!")
}

fn test_scanner_string() {
    var source = "  123   45.67  hello_nora\nnext_line "
    var s = io.NewStrScanner(source)

    // 1. Scan int
    if (!s.Next()) {
        panic("expected first token (int)")
    }
    var val_int_opt = s.Int()
    if (!val_int_opt.IsSome()) {
        panic("expected int token value")
    }
    var val_int = val_int_opt.Unwrap()
    io.PrintLn("Parsed int: ${val_int}")
    if (val_int != 123) {
        panic("int mismatch")
    }

    // 2. Scan float
    if (!s.Next()) {
        panic("expected second token (float)")
    }
    var val_float_opt = s.Float()
    if (!val_float_opt.IsSome()) {
        panic("expected float token value")
    }
    var val_float = val_float_opt.Unwrap()
    io.PrintLn("Parsed float: ${val_float}")
    if (val_float != 45.67) {
        panic("float mismatch")
    }

    // 3. Scan string
    if (!s.Next()) {
        panic("expected third token (string)")
    }
    var val_str = s.Text()
    io.PrintLn("Parsed string: ${val_str}")
    if (val_str != "hello_nora") {
        panic("string mismatch")
    }
    // 4. Scan next line
    if (!s.NextLine()) {
        panic("expected fourth token (next line)")
    }
    var val_line = s.Text()
    io.PrintLn("Parsed line: ${val_line}")
    if (val_line != "next_line ") {
        panic("line mismatch")
    }

    // 5. Scan EOF
    if (s.Next()) {
        panic("expected EOF")
    }
    io.PrintLn("test_scanner_string PASSED!")
}

fn test_scanner_file() {
    var path = "test_scanner_file.txt"
    var content = "789 98.76 file_word\nline_two"

    var write_res = io.WriteFile(path, content)
    if (!write_res.IsOk()) {
        panic("failed to write scanner file")
    }

    var open_read_res = io.OpenFile(path, "r")
    match open_read_res {
        Ok(f) => {
            var s = io.NewFileScanner(f)

            // 1. Scan int
            if (!s.Next()) {
                panic("expected file first token")
            }
            var opt_int = s.Int()
            var val_int = opt_int.Unwrap()
            io.PrintLn("Parsed file int: ${val_int}")
            if (val_int != 789) {
                panic("file int mismatch")
            }

            // 2. Scan float
            if (!s.Next()) {
                panic("expected file second token")
            }
            var opt_float = s.Float()
            var val_float = opt_float.Unwrap()
            io.PrintLn("Parsed file float: ${val_float}")
            if (val_float != 98.76) {
                panic("file float mismatch")
            }

            // 3. Scan string
            if (!s.Next()) {
                panic("expected file third token")
            }
            var val_str = s.Text()
            io.PrintLn("Parsed file string: ${val_str}")
            if (val_str != "file_word") {
                panic("file string mismatch")
            }

            // 4. Scan next line
            if (!s.NextLine()) {
                panic("expected file next line")
            }
            var val_line = s.Text()
            io.PrintLn("Parsed file line: ${val_line}")
            if (val_line != "line_two") {
                panic("file line mismatch")
            }

            // 5. Scan EOF
            if (s.Next()) {
                panic("expected file EOF")
            }
            // f drops/closes here when match block exits!
        }
        Err(err) => {
            panic("failed to open file for scanning")
        }
    }
    io.PrintLn("test_scanner_file PASSED!")
}

fn main() i32 {
    test_file_static_helpers()
    test_file_raii_methods()
    test_scanner_string()
    test_scanner_file()

    io.PrintLn("All I/O & Scanner tests PASSED 100% successfully!")
    return 0
}
