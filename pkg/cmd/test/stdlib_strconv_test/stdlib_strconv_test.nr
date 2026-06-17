package main

import "io"
import "strconv"

fn assert_string_eq(val: #str, expected: #str, name: #str) {
    if (val != expected) {
        io.PrintLn("Assertion failed for: ")
        io.PrintLn(name)
        io.PrintLn("Expected: ")
        io.PrintLn(expected)
        io.PrintLn("Got: ")
        io.PrintLn(val)
        panic("Error: assertion failed")
    }
}

fn assert_int_eq(val: i32, expected: i32, name: #str) {
    if (val != expected) {
        io.PrintLn("Assertion failed for: ")
        io.PrintLn(name)
        io.PrintLn("Expected: ${expected}")
        io.PrintLn("Got: ${val}")
        panic("Error: assertion failed")
    }
}

fn assert_double_approx(val: f64, expected: f64, epsilon: f64, name: #str) {
    var diff = val - expected
    if (diff < 0.0) {
        diff = -diff
    }
    if (diff > epsilon) {
        io.PrintLn("Assertion failed for: ")
        io.PrintLn(name)
        io.PrintLn("Expected: approx ${expected}")
        io.PrintLn("Got: ${val}")
        panic("Error: assertion failed")
    }
}

fn assert_bool_eq(val: bool, expected: bool, name: #str) {
    if (val != expected) {
        io.PrintLn("Assertion failed for: ")
        io.PrintLn(name)
        if (expected) {
            io.PrintLn("Expected: true")
        } else {
            io.PrintLn("Expected: false")
        }
        if (val) {
            io.PrintLn("Got: true")
        } else {
            io.PrintLn("Got: false")
        }
        panic("Error: assertion failed")
    }
}

fn main() {
    io.PrintLn("=== Starting Nora Premium strconv Standard Library Test Suite ===")

    // 1. Integer Parsing
    io.PrintLn("  1. Testing Integer Parsing (ParseInt / Atoi)...")

    match strconv.ParseInt("12345", 10) {
        Ok(v) => {
            assert_int_eq(v, 12345, "ParseInt decimal")
        }
        Err(e) => {
            panic("ParseInt failed: " + e)
        }
    }

    match strconv.ParseInt("-6789", 10) {
        Ok(v) => {
            assert_int_eq(v, -6789, "ParseInt negative decimal")
        }
        Err(e) => {
            panic("ParseInt failed: " + e)
        }
    }

    match strconv.ParseInt("+42", 10) {
        Ok(v) => {
            assert_int_eq(v, 42, "ParseInt positive sign")
        }
        Err(e) => {
            panic("ParseInt failed: " + e)
        }
    }

    match strconv.ParseInt("0x1F", 0) {
        Ok(v) => {
            assert_int_eq(v, 31, "ParseInt hex base-0")
        }
        Err(e) => {
            panic("ParseInt failed: " + e)
        }
    }

    match strconv.ParseInt("0b1010", 0) {
        Ok(v) => {
            assert_int_eq(v, 10, "ParseInt binary base-0")
        }
        Err(e) => {
            panic("ParseInt failed: " + e)
        }
    }

    match strconv.ParseInt("0o75", 0) {
        Ok(v) => {
            assert_int_eq(v, 61, "ParseInt octal base-0")
        }
        Err(e) => {
            panic("ParseInt failed: " + e)
        }
    }

    match strconv.ParseInt("2147483647", 10) {
        Ok(v) => {
            assert_int_eq(v, 2147483647, "ParseInt max int")
        }
        Err(e) => {
            panic("ParseInt failed: " + e)
        }
    }

    // Overflow check
    match strconv.ParseInt("2147483648", 10) {
        Ok(v) => {
            panic("Expected ParseInt overflow error but succeeded")
        }
        Err(e) => {
            assert_string_eq(e, "value out of range", "ParseInt overflow validation")
        }
    }

    // Syntax validation
    match strconv.ParseInt("123a45", 10) {
        Ok(v) => {
            panic("Expected ParseInt syntax error but succeeded")
        }
        Err(e) => {
            assert_string_eq(e, "invalid syntax", "ParseInt syntax validation")
        }
    }

    match strconv.Atoi("98765") {
        Ok(v) => {
            assert_int_eq(v, 98765, "Atoi")
        }
        Err(e) => {
            panic("Atoi failed: " + e)
        }
    }

    // 2. Integer Formatting
    io.PrintLn("  2. Testing Integer Formatting (FormatInt / Itoa)...")

    var s_f1 = strconv.FormatInt(12345, 10)
    assert_string_eq(s_f1, "12345", "FormatInt base 10")
    assert_string_eq(strconv.FormatInt(12345, 10), "12345", "FormatInt base 10 inline")

    var s_f2 = strconv.FormatInt(-6789, 10)
    assert_string_eq(s_f2, "-6789", "FormatInt negative base 10")

    var s_f3 = strconv.FormatInt(31, 16)
    assert_string_eq(s_f3, "1f", "FormatInt base 16")

    var s_f4 = strconv.FormatInt(10, 2)
    assert_string_eq(s_f4, "1010", "FormatInt base 2")

    var s_f5 = strconv.FormatInt(61, 8)
    assert_string_eq(s_f5, "75", "FormatInt base 8")

    var s_f6 = strconv.Itoa(98765)
    assert_string_eq(s_f6, "98765", "Itoa")

    var s_f7 = strconv.FormatInt(-2147483648, 10)
    assert_string_eq(s_f7, "-2147483648", "FormatInt INT_MIN")

    // 3. Float Parsing
    io.PrintLn("  3. Testing Float Parsing (ParseFloat)...")

    match strconv.ParseFloat("3.14159") {
        Ok(v) => {
            assert_double_approx(v, 3.14159, 0.00001, "ParseFloat simple")
        }
        Err(e) => {
            panic("ParseFloat failed: " + e)
        }
    }

    match strconv.ParseFloat("-0.0123") {
        Ok(v) => {
            assert_double_approx(v, -0.0123, 0.0001, "ParseFloat negative")
        }
        Err(e) => {
            panic("ParseFloat failed: " + e)
        }
    }

    match strconv.ParseFloat("1.23e4") {
        Ok(v) => {
            assert_double_approx(v, 12300.0, 0.1, "ParseFloat exponent")
        }
        Err(e) => {
            panic("ParseFloat failed: " + e)
        }
    }

    match strconv.ParseFloat("invalid_float") {
        Ok(v) => {
            panic("Expected ParseFloat syntax error but succeeded")
        }
        Err(e) => {
            assert_string_eq(e, "invalid syntax", "ParseFloat syntax validation")
        }
    }

    // 4. Float Formatting
    io.PrintLn("  4. Testing Float Formatting (FormatFloat)...")

    var s_ff1 = strconv.FormatFloat(3.14159, "f", 2)
    assert_string_eq(s_ff1, "3.14", "FormatFloat f 2")

    var s_ff2 = strconv.FormatFloat(3.14159, "f", 4)
    assert_string_eq(s_ff2, "3.1416", "FormatFloat f 4")

    var s_ff3 = strconv.FormatFloat(3.14, "g", -1)
    assert_string_eq(s_ff3, "3.14", "FormatFloat g")

    // 5. Boolean Conversion
    io.PrintLn("  5. Testing Boolean Conversion (ParseBool / FormatBool)...")

    match strconv.ParseBool("true") {
        Ok(v) => {
            assert_bool_eq(v, true, "ParseBool true")
        }
        Err(e) => {
            panic("ParseBool failed: " + e)
        }
    }
    match strconv.ParseBool("0") {
        Ok(v) => {
            assert_bool_eq(v, false, "ParseBool false")
        }
        Err(e) => {
            panic("ParseBool failed: " + e)
        }
    }
    match strconv.ParseBool("invalid") {
        Ok(v) => {
            panic("Expected ParseBool error but succeeded")
        }
        Err(e) => {
            assert_string_eq(e, "invalid syntax", "ParseBool syntax validation")
        }
    }

    var s_fb1 = strconv.FormatBool(true)
    assert_string_eq(s_fb1, "true", "FormatBool true")

    var s_fb2 = strconv.FormatBool(false)
    assert_string_eq(s_fb2, "false", "FormatBool false")

    // 6. Quoting & Escaping
    io.PrintLn("  6. Testing String Quoting (Quote / Unquote)...")

    var orig = "Hello \"World\"!\n\t"
    var quoted = strconv.Quote(orig)
    assert_string_eq(quoted, "\"Hello \\\"World\\\"!\\n\\t\"", "Quote encoding")

    match strconv.Unquote(quoted) {
        Ok(unquoted) => {
            assert_string_eq(unquoted, orig, "Unquote decoding roundtrip")
        }
        Err(e) => {
            panic("Unquote failed: " + e)
        }
    }

    // Hex escape decode test
    match strconv.Unquote("\"\\x4e\\x6f\\x72\\x61\"") {
        Ok(unquoted) => {
            assert_string_eq(unquoted, "Nora", "Unquote hex bytes")
        }
        Err(e) => {
            panic("Unquote hex failed: " + e)
        }
    }

    // 7. High-Performance Buffer Appenders
    io.PrintLn("  7. Testing Buffer Appenders (Append*)...")

    var buf0 = alloc byte[0]
    var buf1 = strconv.AppendInt(buf0, 100, 10)

    // Inline bytes-to-string conversion
    var buf_len1 = len(buf1)
    var s1 = alloc byte[buf_len1 + 1]
    var i1 = 0
    while (i1 < buf_len1) {
        s1[i1] = buf1[i1]
        i1 = i1 + 1
    }
    s1[buf_len1] = 0
    assert_string_eq(s1, "100", "AppendInt")

    var buf2 = strconv.AppendBool(buf1, true)
    var buf_len2 = len(buf2)
    var s2 = alloc byte[buf_len2 + 1]
    var i2 = 0
    while (i2 < buf_len2) {
        s2[i2] = buf2[i2]
        i2 = i2 + 1
    }
    s2[buf_len2] = 0
    assert_string_eq(s2, "100true", "AppendBool")

    var buf3 = strconv.AppendFloat(buf2, 3.14, "f", 2)
    var buf_len3 = len(buf3)
    var s3 = alloc byte[buf_len3 + 1]
    var i3 = 0
    while (i3 < buf_len3) {
        s3[i3] = buf3[i3]
        i3 = i3 + 1
    }
    s3[buf_len3] = 0
    assert_string_eq(s3, "100true3.14", "AppendFloat")

    var buf4 = strconv.AppendQuote(buf3, "ok")
    var buf_len4 = len(buf4)
    var s4 = alloc byte[buf_len4 + 1]
    var i4 = 0
    while (i4 < buf_len4) {
        s4[i4] = buf4[i4]
        i4 = i4 + 1
    }
    s4[buf_len4] = 0
    assert_string_eq(s4, "100true3.14\"ok\"", "AppendQuote")

    io.PrintLn("=== All Standard strconv Conversions Tests Passed Successfully! ===")
}
