package io

// ─── C stdlib bindings ────────────────────────────────────────────────────────
extern fn snprintf(buf: ptr, size: i32, format: str, ...) i32
extern fn getchar() i32
extern fn fopen(path: str, mode: str) ptr
extern fn fclose(file: ptr) i32
extern fn fputs(s: str, file: ptr) i32
extern fn fputc(c: i32, file: ptr) i32
extern fn fgetc(file: ptr) i32
extern fn fgets(buf: ptr, size: i32, file: ptr) ptr
extern fn fread(buf: ptr, elem_size: i32, count: i32, file: ptr) i32
extern fn fwrite(buf: ptr, elem_size: i32, count: i32, file: ptr) i32
extern fn fflush(file: ptr) i32
extern fn feof(file: ptr) i32
extern fn ferror(file: ptr) i32
extern fn rewind(file: ptr)
extern fn fseek(file: ptr, offset: i32, whence: i32) i32
extern fn ftell(file: ptr) i32

// FILE* accessors (stdin/stdout/stderr are C macros, not linkable symbols)
extern fn nr_get_stdout() ptr
extern fn nr_get_stderr() ptr
extern fn nr_get_stdin() ptr

// ─── COMPILER INTRINSICS (Poly-Builtins) ──────────────────────────────────────
// These functions are intercepted by the compiler to support variadic types
// and zero-allocation string interpolation.

[macro("io_macro", "expand_print")]
export fn print()

[macro("io_macro", "expand_println")]
export fn println()

[macro("io_macro", "expand_eprint")]
fn eprint()

[macro("io_macro", "expand_eprintln")]
fn eprintln()

[macro("io_macro", "expand_scan")]
fn scan()

[macro("io_macro", "expand_scanln")]
fn scanln()
// ─────────────────────────────────────────────────────────────────────────────

// ─── Formatting (value → string) ──────────────────────────────────────────────

// format_int converts an integer to its string representation.
fn format_int(n: i32) str {
    var buf = alloc byte[32]
    snprintf(buf, 32, "%d", n)
    return buf
}

// format_float converts a float to its string representation.
fn format_float(f: f64) str {
    var buf = alloc byte[64]
    snprintf(buf, 64, "%g", f)
    return buf
}

// format_bool converts a bool to "true" or "false".
fn format_bool(b: bool) str {
    if (b) {
        return "true"
    }
    return "false"
}

// ─── Standard Input ───────────────────────────────────────────────────────────

// read_line reads one line from stdin (strips the trailing newline).
// Allocates up to 4095 bytes. Returns an empty string on EOF.
fn read_line() str {
    var buf = alloc byte[4096]
    var i = 0
    var c = getchar()
    while c != 10 && c != -1 && i < 4095 {
        buf[i] = c
        i = i + 1
        c = getchar()
    }
    buf[i] = 0
    return buf
}

// read_char reads a single character from stdin and returns its ASCII value.
// Returns -1 on EOF.
fn read_char() i32 {
    return getchar()
}

// read_int reads a decimal integer from stdin (skips leading whitespace).
fn read_int() i32 {
    var c = getchar()
    var result = 0
    var negative = false
    // Skip whitespace
    while c == 32 || c == 10 || c == 9 || c == 13 {
        c = getchar()
    }
    // Optional minus sign
    if (c == 45) {
        negative = true
        c = getchar()
    }
    // Accumulate digits
    while c >= 48 && c <= 57 {
        result = result * 10 + (c - 48)
        c = getchar()
    }
    if (negative) {
        return 0 - result
    }
    return result
}

// scan_str reads a whitespace-delimited word from stdin.
// Allocates up to 4095 bytes. Returns an empty string on EOF.
fn scan_str() str {
    var buf = alloc byte[4096]
    var i = 0
    var c = getchar()
    // Skip leading whitespace
    while c == 32 || c == 10 || c == 9 || c == 13 {
        c = getchar()
    }
    // Read characters until whitespace or EOF
    while c != 32 && c != 10 && c != 9 && c != 13 && c != -1 && i < 4095 {
        buf[i] = c
        i = i + 1
        c = getchar()
    }
    buf[i] = 0
    return buf
}

// ─── File I/O ─────────────────────────────────────────────────────────────────

// Seek-whence constants (mirror SEEK_SET / SEEK_CUR / SEEK_END from <stdio.h>)
fn seek_set() i32 { return 0 }
fn seek_cur() i32 { return 1 }
fn seek_end() i32 { return 2 }

// open opens a file. Mode: "r", "w", "a", "rb", "wb", etc.
// Returns none on failure.
fn open(path: str, mode: str) ptr {
    return fopen(path, mode)
}

// close closes an open file handle.
fn close(f: ptr) {
    fclose(f)
}

// write_to writes a string to a file (no newline).
fn write_to(f: ptr, s: str) {
    fputs(s, f)
}

// writeln_to writes a string followed by a newline to a file.
fn writeln_to(f: ptr, s: str) {
    fputs(s, f)
    fputc(10, f)
}

// write_char_to writes a single character (ASCII value) to a file.
fn write_char_to(f: ptr, c: i32) {
    fputc(c, f)
}

// read_line_from reads one line from a file (strips the newline).
// Returns none at end of file.
fn read_line_from(f: ptr) str {
    if (is_eof(f)) {
        return none
    }
    var buf = alloc byte[4096]
    var i = 0
    var c = fgetc(f)
    while c != 10 && c != -1 && i < 4095 {
        buf[i] = c
        i = i + 1
        c = fgetc(f)
    }
    if (i == 0 && c == -1) {
        return none
    }
    buf[i] = 0
    return buf
}

// is_eof returns true when the file is at end-of-file.
fn is_eof(f: ptr) bool {
    return feof(f) != 0
}

// has_error returns true when the file's error indicator is set.
fn has_error(f: ptr) bool {
    return ferror(f) != 0
}

// flush flushes a file's write buffer.
fn flush(f: ptr) {
    fflush(f)
}

// flush_stdout flushes stdout (useful before reading input).
fn flush_stdout() {
    fflush(nr_get_stdout())
}

// seek moves the file position. Use seek_set(), seek_cur(), seek_end() for whence.
// Returns true on success.
fn seek(f: ptr, offset: i32, whence: i32) bool {
    return fseek(f, offset, whence) == 0
}

// tell returns the current byte position in the file.
fn tell(f: ptr) i32 {
    return ftell(f)
}

// rewind_file moves the file position back to the beginning.
fn rewind_file(f: ptr) {
    rewind(f)
}
