package io

extern fn nr_i32_to_str(v: i32) str

// ─── C stdlib bindings ────────────────────────────────────────────────────────
extern fn snprintf(buf: ptr, size: i32, format: #str, ...) i32

extern fn getchar() i32

extern fn fopen(path: #str, mode: #str) ptr

extern fn fclose(file: ptr) i32

extern fn fputs(s: #str, file: ptr) i32

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

extern fn strlen(s: #str) i32

extern fn nr_atoi(s: #str) i32

extern fn nr_atof(s: #str) f64

extern fn isspace(c: i32) i32

// FILE* accessors (stdin/stdout/stderr are C macros, not linkable symbols)
extern fn nr_get_stdout() ptr

extern fn nr_get_stderr() ptr

extern fn nr_get_stdin() ptr

// ─── COMPILER INTRINSICS (Poly-Builtins) ──────────────────────────────────────
// These functions are intercepted by the compiler to support variadic types
// and zero-allocation string interpolation.
[macro("io_macro", "expand_print")]
pub fn Print()

[macro("io_macro", "expand_println")]
pub fn PrintLn()

[macro("io_macro", "expand_eprint")]
pub fn EPrint()

[macro("io_macro", "expand_eprintln")]
pub fn EprintLn()

[macro("io_macro", "expand_scan")]
pub fn Scan()

[macro("io_macro", "expand_scanln")]
pub fn ScanLn()

// ─────────────────────────────────────────────────────────────────────────────
// ─── Formatting (value → string) ──────────────────────────────────────────────
// FormatInt converts an integer to its string representation.
pub fn FormatInt(n: i32) str {
    var buf = alloc byte[32]
    snprintf(buf, 32, "%d", n)
    return buf
}

// FormatFloat converts a float to its string representation.
pub fn FormatFloat(f: f64) str {
    var buf = alloc byte[64]
    snprintf(buf, 64, "%g", f)
    return buf
}

// FormatBool converts a bool to "true" or "false".
pub fn FormatBool(b: bool) str {
    if (b) {
        return "true"
    }
    return "false"
}

// ─── Standard Input ───────────────────────────────────────────────────────────
// ReadLine reads one line from stdin (strips the trailing newline).
// Allocates up to 4095 bytes. Returns an empty string on EOF.
pub fn ReadLine() str {
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

// ReadChar reads a single character from stdin and returns its ASCII value.
// Returns -1 on EOF.
pub fn ReadChar() i32 {
    return getchar()
}

// ReadInt reads a decimal integer from stdin (skips leading whitespace).
pub fn ReadInt() i32 {
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

// ScanStr reads a whitespace-delimited word from stdin.
// Allocates up to 4095 bytes. Returns an empty string on EOF.
pub fn ScanStr() str {
    var buf = alloc byte[4096]
    var i = 0
    var c = getchar()
    // Skip leading whitespace
    while c == 32 || c == 10 || c == 9 || c == 13 {
        c = getchar()
    }
    // Read characters until whitespace or EOF
    while c != 32 && c != 10 && c != 9 && c != 13 && c != -1 &&
        i < 4095 {
        buf[i] = c
        i = i + 1
        c = getchar()
    }
    buf[i] = 0
    return buf
}

// Seek-whence constants (mirror SEEK_SET / SEEK_CUR / SEEK_END from <stdio.h>)
pub fn SeekSet() i32 {
    return 0
}

pub fn SeekCur() i32 {
    return 1
}

pub fn SeekEnd() i32 {
    return 2
}

pub type File = struct {
    handle: ptr
}

// Destructor/Drop routine
pub fn (self: &File) drop() {
    if (self.handle != none) {
        fclose(self.handle)
        self.handle = none
    }
}

pub fn (self: &File) IsOpen() bool {
    return self.handle != none
}

pub fn (self: &File) ReadLine() Option[str] {
    if (self.handle == none) {
        return None[str]
    }
    if (feof(self.handle) != 0) {
        return None[str]
    }
    var buf = alloc byte[4096]
    var i = 0
    var c = fgetc(self.handle)
    while c != 10 && c != -1 && i < 4095 {
        buf[i] = c
        i = i + 1
        c = fgetc(self.handle)
    }
    if (i == 0 && c == -1) {
        return None[str]
    }
    buf[i] = 0
    return Some[str](buf)
}

pub fn (self: &File) ReadChar() i32 {
    if (self.handle == none) {
        return -1
    }
    return fgetc(self.handle)
}

pub fn (self: &File) Write(content: #str) bool {
    if (self.handle == none) {
        return false
    }
    return fputs(content, self.handle) >= 0
}

pub fn (self: &File) Writeln(content: #str) bool {
    if (self.handle == none) {
        return false
    }
    fputs(content, self.handle)
    return fputc(10, self.handle) >= 0
}

pub fn (self: &File) WriteChar(c: i32) {
    if (self.handle != none) {
        fputc(c, self.handle)
    }
}

pub fn (self: &File) Seek(offset: i32, whence: i32) bool {
    if (self.handle == none) {
        return false
    }
    return fseek(self.handle, offset, whence) == 0
}

pub fn (self: &File) Tell() i32 {
    if (self.handle == none) {
        return -1
    }
    return ftell(self.handle)
}

pub fn (self: &File) Rewind() {
    if (self.handle != none) {
        rewind(self.handle)
    }
}

pub fn (self: &File) Flush() {
    if (self.handle != none) {
        fflush(self.handle)
    }
}

pub fn (self: &File) IsEOF() bool {
    if (self.handle == none) {
        return true
    }
    return feof(self.handle) != 0
}

pub fn (self: &File) HasError() bool {
    if (self.handle == none) {
        return true
    }
    return ferror(self.handle) != 0
}

// ─── Static High-Level File Helpers ───────────────────────────────────────────
pub fn OpenFile(path: #str, mode: #str) Result[File, str] {
    var h = fopen(path, mode)
    if (h == none) {
        return Err[File, str]("failed to open file: " + path)
    }
    return Ok[File, str](File {
        handle: h
    })
}

pub fn ReadFile(path: #str) Result[str, str] {
    var f = OpenFile(path, "r")?
    var f_raw = f.handle

    fseek(f_raw, 0, 2) // SEEK_END
    var size = ftell(f_raw)
    fseek(f_raw, 0, 0) // SEEK_SET

    if (size < 0) {
        return Err[str, str]("failed to determine file size")
    }

    var buf = alloc byte[size + 1]
    var read_bytes = fread(buf, 1, size, f_raw)
    buf[read_bytes] = 0
    return Ok[str, str](buf)
}

pub fn WriteFile(path: #str, content: #str) Result[bool, str] {
    var f = OpenFile(path, "w")?
    f.Write(content)
    return Ok[bool, str](true)
}

// ─── Dynamic Input Scanner ───────────────────────────────────────────────────
pub type Scanner = struct {
    pos: i32,
    len: i32,
    source_str: str,
    source_file: ptr,
    current_token: str
}

pub fn NewStrScanner(source: str) Scanner {
    return Scanner {
        source_str: source,
        source_file: none,
        pos: 0,
        len: strlen(source),
        current_token: ""
    }
}

pub fn NewFileScanner(f: &File) Scanner {
    return Scanner {
        source_str: "",
        source_file: f.handle,
        pos: 0,
        len: 0,
        current_token: ""
    }
}

pub fn NewStdinScanner() Scanner {
    return Scanner {
        source_str: "",
        source_file: nr_get_stdin(),
        pos: 0,
        len: 0,
        current_token: ""
    }
}

[unsafe]
pub fn (self: &Scanner) Next() bool {
    if (self.source_file != none) {
        var c = fgetc(self.source_file)
        while c != -1 && (c == 32 || c == 9 || c == 10 ||
            c == 13) {
            c = fgetc(self.source_file)
        }
        if (c == -1) {
            return false
        }
        self.current_token = alloc byte[4096]
        var i = 0
        while c != -1 && c != 32 && c != 9 && c != 10 &&
            c != 13 && i < 4095 {
            self.current_token[i] = c
            i = i + 1
            c = fgetc(self.source_file)
        }
        self.current_token[i] = 0
        return true
    } else {
        var start = self.pos
        while start < self.len && (self.source_str[start] == 32 ||
            self.source_str[start] == 9 || self.source_str[start] == 10 ||
            self.source_str[start] == 13) {
            start = start + 1
        }
        if (start >= self.len) {
            self.pos = start
            return false
        }
        var end = start
        while end < self.len && self.source_str[end] != 32 &&
            self.source_str[end] != 9 && self.source_str[end] != 10 &&
            self.source_str[end] != 13 {
            end = end + 1
        }
        var t_len = end - start
        self.current_token = alloc byte[t_len + 1]
        var j = 0
        while j < t_len {
            self.current_token[j] = self.source_str[start + j]
            j = j + 1
        }
        self.current_token[t_len] = 0
        self.pos = end
        return true
    }
}

[unsafe]
pub fn (self: &Scanner) NextLine() bool {
    if (self.source_file != none) {
        if (feof(self.source_file) != 0) {
            return false
        }
        var c = fgetc(self.source_file)
        while c == 10 || c == 13 {
            c = fgetc(self.source_file)
        }
        if (c == -1) {
            return false
        }
        self.current_token = alloc byte[4096]
        var i = 0
        while c != 10 && c != 13 && c != -1 && i < 4095 {
            self.current_token[i] = c
            i = i + 1
            c = fgetc(self.source_file)
        }
        self.current_token[i] = 0
        return true
    } else {
        var start = self.pos
        while start < self.len && (self.source_str[start] == 10 ||
            self.source_str[start] == 13) {
            start = start + 1
        }
        if (start >= self.len) {
            return false
        }
        var end = start
        while end < self.len && self.source_str[end] != 10 &&
            self.source_str[end] != 13 {
            end = end + 1
        }
        var t_len = end - start
        self.current_token = alloc byte[t_len + 1]
        var j = 0
        while j < t_len {
            self.current_token[j] = self.source_str[start + j]
            j = j + 1
        }
        self.current_token[t_len] = 0

        if (end < self.len && self.source_str[end] == 13) {
            end = end + 1
        }
        if (end < self.len && self.source_str[end] == 10) {
            end = end + 1
        }
        self.pos = end
        return true
    }
}

pub fn (self: &Scanner) Text() #str {
    return self.current_token
}

pub fn (self: &Scanner) Int() Option[i32] {
    if (strlen(self.current_token) == 0) {
        return None[i32]
    }
    return Some[i32](nr_atoi(self.current_token))
}

pub fn (self: &Scanner) Float() Option[f64] {
    if (strlen(self.current_token) == 0) {
        return None[f64]
    }
    return Some[f64](nr_atof(self.current_token))
}
