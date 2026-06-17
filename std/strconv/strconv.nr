package strconv

import "string"

extern fn strlen(s: #str) i32

extern fn nr_atof(s: #str) f64

extern fn snprintf(buf: ptr, size: i32, format: #str, ...) i32

// ParseInt parses a string in the given base (2 to 36) and returns the i32 value.
// If base is 0, it automatically detects the base from the string prefix:
// "0x" or "0X" for base 16, "0b" or "0B" for base 2, "0o" or "0O" for base 8,
// and base 10 otherwise.
pub fn ParseInt(s: #str, base: i32) Result[i32, str] {
    var len_s = 0
    while (s[len_s] != 0) {
        len_s = len_s + 1
    }
    if (len_s == 0) {
        return Err[i32, str]("empty string")
    }

    var neg = false
    var start = 0
    if (s[0] == 45) { // '-'
        neg = true
        start = 1
    } else if (s[0] == 43) { // '+'
        start = 1
    }

    if (start >= len_s) {
        return Err[i32, str]("invalid syntax")
    }

    var actual_base = base
    if (base == 0) {
        if (start + 2 <= len_s && s[start] == 48) { // '0'
            var c2 = s[start + 1]
            if (c2 == 120 || c2 == 88) { // 'x' or 'X'
                actual_base = 16
                start = start + 2
            } else if (c2 == 98 || c2 == 66) { // 'b' or 'B'
                actual_base = 2
                start = start + 2
            } else if (c2 == 111 || c2 == 79) { // 'o' or 'O'
                actual_base = 8
                start = start + 2
            } else {
                actual_base = 8
                start = start + 1
            }
        } else {
            actual_base = 10
        }
    } else if (base == 16) {
        if (start + 2 <= len_s && s[start] == 48 && (s[start + 1] == 120 ||
            s[start + 1] == 88)) {
            start = start + 2
        }
    }

    if (actual_base < 2 || actual_base > 36) {
        return Err[i32, str]("invalid base")
    }

    if (start >= len_s) {
        return Err[i32, str]("invalid syntax")
    }

    var res = 0
    var i = start
    while (i < len_s) {
        var c = s[i]
        var digit = -1
        if (c >= 48 && c <= 57) { // '0' - '9'
            digit = c - 48
        } else if (c >= 97 && c <= 122) { // 'a' - 'z'
            digit = c - 97 + 10
        } else if (c >= 65 && c <= 90) { // 'A' - 'Z'
            digit = c - 65 + 10
        }

        if (digit < 0 || digit >= actual_base) {
            return Err[i32, str]("invalid syntax")
        }

        // Check overflow (i32 max is 2147483647)
        if (res > (2147483647 - digit) / actual_base) {
            return Err[i32, str]("value out of range")
        }

        res = res * actual_base + digit
        i = i + 1
    }

    if (neg) {
        return Ok[i32, str](-res)
    }
    return Ok[i32, str](res)
}

// Atoi parses a base-10 integer string.
pub fn Atoi(s: #str) Result[i32, str] {
    var res = ParseInt(s, 10)
    return res
}

// FormatInt converts an i32 to a string representation in the specified base (2 to 36).
pub fn FormatInt(val: i32, base: i32) str {
    if (base < 2 || base > 36) {
        return ""
    }
    if (val == 0) {
        return "0"
    }

    var is_neg = false
    var uval = val
    if (val < 0) {
        is_neg = true
        if (val == -2147483648) {
            if (base == 10) {
                return "-2147483648"
            }
            if (base == 16) {
                return "-80000000"
            }
            if (base == 2) {
                return "-10000000000000000000000000000000"
            }
            if (base == 8) {
                return "-20000000000"
            }
        }
        uval = -val
    }

    var chars = "0123456789abcdefghijklmnopqrstuvwxyz"
    var buf = alloc byte[65]
    var pos = 64
    buf[pos] = 0

    while (uval > 0) {
        pos = pos - 1
        var rem = uval % base
        buf[pos] = chars[rem]
        uval = uval / base
    }

    if (is_neg) {
        pos = pos - 1
        buf[pos] = 45 // '-'
    }

    var length = 64 - pos
    var res = alloc byte[length + 1]
    var i = 0
    while (i < length) {
        res[i] = buf[pos + i]
        i = i + 1
    }
    res[length] = 0
    return res
}

// Itoa converts an integer to its base-10 string representation.
pub fn Itoa(val: i32) str {
    var res = FormatInt(val, 10)
    return res
}

// ParseFloat parses a floating-point string and returns the f64 value.
pub fn ParseFloat(s: #str) Result[f64, str] {
    var len_s = 0
    while (s[len_s] != 0) {
        len_s = len_s + 1
    }

    var start = 0
    while (start < len_s && (s[start] == 32 || s[start] == 9 ||
        s[start] == 10 || s[start] == 13)) {
        start = start + 1
    }

    var end = len_s - 1
    while (end >= start && (s[end] == 32 || s[end] == 9 ||
        s[end] == 10 || s[end] == 13)) {
        end = end - 1
    }

    if (start > end) {
        return Err[f64, str]("empty string")
    }

    var has_digit = false
    var dec_count = 0
    var exp_count = 0
    var i = start
    while (i <= end) {
        var c = s[i]
        if (c >= 48 && c <= 57) { // '0'-'9'
            has_digit = true
        } else if (c == 45 || c == 43) { // '-' or '+'
            if (i != start && s[i - 1] != 101 && s[i - 1] != 69) {
                return Err[f64, str]("invalid syntax")
            }
        } else if (c == 46) { // '.'
            dec_count = dec_count + 1
            if (dec_count > 1 || exp_count > 0) {
                return Err[f64, str]("invalid syntax")
            }
        } else if (c == 101 || c == 69) { // 'e' or 'E'
            exp_count = exp_count + 1
            if (exp_count > 1 || !has_digit) {
                return Err[f64, str]("invalid syntax")
            }
            if (i == end) {
                return Err[f64, str]("invalid syntax")
            }
        } else {
            return Err[f64, str]("invalid syntax")
        }
        i = i + 1
    }

    if (!has_digit) {
        return Err[f64, str]("invalid syntax")
    }

    var sub_len = end - start + 1
    var tmp = alloc byte[sub_len + 1]
    var j = 0
    while (j < sub_len) {
        tmp[j] = s[start + j]
        j = j + 1
    }
    tmp[sub_len] = 0

    var val = nr_atof(tmp)
    return Ok[f64, str](val)
}

// FormatFloat converts a float to a string according to format fmt ('f', 'e', 'g') and precision.
pub fn FormatFloat(val: f64, fmt: #str, prec: i32) str {
    var format = alloc byte[16]
    if (prec < 0) {
        format[0] = 37 // '%'
        if (fmt[0] == 101 || fmt[0] == 69) {
            format[1] = fmt[0] // 'e' or 'E'
        } else {
            format[1] = 103 // 'g'
        }
        format[2] = 0
    } else {
        format[0] = 37 // '%'
        format[1] = 46 // '.'

        var p_str = Itoa(prec)
        var p_len = strlen(p_str)
        var i = 0
        while (i < p_len) {
            format[2 + i] = p_str[i]
            i = i + 1
        }

        var f_char = 102 // 'f'
        if (fmt[0] == 101 || fmt[0] == 69 || fmt[0] == 103 ||
            fmt[0] == 71) {
            f_char = fmt[0]
        }
        format[2 + p_len] = f_char
        format[2 + p_len + 1] = 0
    }

    var buf = alloc byte[128]
    snprintf(buf, 128, format, val)
    return buf
}

// ParseBool parses a boolean string.
pub fn ParseBool(s: #str) Result[bool, str] {
    if (s == "1" || s == "t" || s == "T" || s == "true" ||
        s == "TRUE" || s == "True") {
        return Ok[bool, str](true)
    }
    if (s == "0" || s == "f" || s == "F" || s == "false" ||
        s == "FALSE" || s == "False") {
        return Ok[bool, str](false)
    }
    return Err[bool, str]("invalid syntax")
}

// FormatBool converts a boolean to its string representation.
pub fn FormatBool(b: bool) str {
    if (b) {
        return "true"
    }
    return "false"
}

// Quote returns a double-quoted string with escape sequences.
pub fn Quote(s: #str) str {
    var len_s = 0
    while (s[len_s] != 0) {
        len_s = len_s + 1
    }

    var buf = alloc byte[2 * len_s + 3]
    buf[0] = 34 // '"'
    var pos = 1

    var i = 0
    while (i < len_s) {
        var c = s[i]
        if (c == 34) { // '"'
            buf[pos] = 92 // '\\'
            buf[pos + 1] = 34
            pos = pos + 2
        } else if (c == 92) { // '\\'
            buf[pos] = 92
            buf[pos + 1] = 92
            pos = pos + 2
        } else if (c == 10) { // '\n'
            buf[pos] = 92
            buf[pos + 1] = 110
            pos = pos + 2
        } else if (c == 13) { // '\r'
            buf[pos] = 92
            buf[pos + 1] = 114
            pos = pos + 2
        } else if (c == 9) { // '\t'
            buf[pos] = 92
            buf[pos + 1] = 116
            pos = pos + 2
        } else if (c < i8(32) || c > i8(126)) {
            buf[pos] = 92
            buf[pos + 1] = 120

            var hex = "0123456789abcdef"
            var h1 = (c >> i8(4)) & i8(15)
            var h2 = c & i8(15)
            buf[pos + 2] = hex[h1]
            buf[pos + 3] = hex[h2]
            pos = pos + 4
        } else {
            buf[pos] = c
            pos = pos + 1
        }
        i = i + 1
    }
    buf[pos] = 34 // '"'
    buf[pos + 1] = 0

    var res = alloc byte[pos + 2]
    var j = 0
    while (j <= pos) {
        res[j] = buf[j]
        j = j + 1
    }
    res[pos + 1] = 0
    return res
}

// Unquote decodes a double-quoted string with escape sequences.
pub fn Unquote(s: #str) Result[str, str] {
    var len_s = 0
    while (s[len_s] != 0) {
        len_s = len_s + 1
    }

    if (len_s < 2 || s[0] != 34 || s[len_s - 1] != 34) {
        return Err[str, str]("invalid syntax")
    }

    var buf = alloc byte[len_s]
    var pos = 0
    var i = 1
    var end = len_s - 1
    while (i < end) {
        var c = s[i]
        if (c == 92) { // '\\'
            if (i + 1 >= end) {
                return Err[str, str]("invalid syntax")
            }
            var next = s[i + 1]
            if (next == 110) { // 'n'
                buf[pos] = 10
                pos = pos + 1
                i = i + 2
            } else if (next == 114) { // 'r'
                buf[pos] = 13
                pos = pos + 1
                i = i + 2
            } else if (next == 116) { // 't'
                buf[pos] = 9
                pos = pos + 1
                i = i + 2
            } else if (next == 34) { // '"'
                buf[pos] = 34
                pos = pos + 1
                i = i + 2
            } else if (next == 92) { // '\\'
                buf[pos] = 92
                pos = pos + 1
                i = i + 2
            } else if (next == 120) { // 'x'
                if (i + 3 >= end) {
                    return Err[str, str]("invalid syntax")
                }

                var h1_char = s[i + 2]
                var h2_char = s[i + 3]
                var h1 = -1
                var h2 = -1

                if (h1_char >= 48 && h1_char <= 57) {
                    h1 = h1_char - 48
                } else if (h1_char >= 97 && h1_char <= 102) {
                    h1 = h1_char - 97 + 10
                } else if (h1_char >= 65 && h1_char <= 70) {
                    h1 = h1_char - 65 + 10
                }

                if (h2_char >= 48 && h2_char <= 57) {
                    h2 = h2_char - 48
                } else if (h2_char >= 97 && h2_char <= 102) {
                    h2 = h2_char - 97 + 10
                } else if (h2_char >= 65 && h2_char <= 70) {
                    h2 = h2_char - 65 + 10
                }

                if (h1 < 0 || h2 < 0) {
                    return Err[str, str]("invalid syntax")
                }

                buf[pos] = (h1 << 4) | h2
                pos = pos + 1
                i = i + 4
            } else {
                return Err[str, str]("invalid syntax")
            }
        } else {
            buf[pos] = c
            pos = pos + 1
            i = i + 1
        }
    }
    buf[pos] = 0

    var res = alloc byte[pos + 1]
    var j = 0
    while (j < pos) {
        res[j] = buf[j]
        j = j + 1
    }
    res[pos] = 0
    return Ok[str, str](res)
}

// AppendInt appends the string representation of val to dst and returns the extended buffer.
pub fn AppendInt(dst: #byte[0], val: i32, base: i32) byte[0] {
    var s = FormatInt(val, base)
    var s_len = strlen(s)
    var dst_len = len(dst)
    var res = alloc byte[dst_len + s_len]

    var i = 0
    while (i < dst_len) {
        res[i] = dst[i]
        i = i + 1
    }

    var j = 0
    while (j < s_len) {
        res[dst_len + j] = s[j]
        j = j + 1
    }
    return res
}

// AppendBool appends "true" or "false" to dst and returns the extended buffer.
pub fn AppendBool(dst: #byte[0], b: bool) byte[0] {
    var s = FormatBool(b)
    var s_len = strlen(s)
    var dst_len = len(dst)
    var res = alloc byte[dst_len + s_len]

    var i = 0
    while (i < dst_len) {
        res[i] = dst[i]
        i = i + 1
    }

    var j = 0
    while (j < s_len) {
        res[dst_len + j] = s[j]
        j = j + 1
    }
    return res
}

// AppendFloat appends the float representation to dst and returns the extended buffer.
pub fn AppendFloat(dst: #byte[0], val: f64, fmt: #str, prec: i32) byte[0] {
    var s = FormatFloat(val, fmt, prec)
    var s_len = strlen(s)
    var dst_len = len(dst)
    var res = alloc byte[dst_len + s_len]

    var i = 0
    while (i < dst_len) {
        res[i] = dst[i]
        i = i + 1
    }

    var j = 0
    while (j < s_len) {
        res[dst_len + j] = s[j]
        j = j + 1
    }
    return res
}

// AppendQuote appends the quoted representation of s to dst and returns the extended buffer.
pub fn AppendQuote(dst: #byte[0], s: #str) byte[0] {
    var q = Quote(s)
    var q_len = strlen(q)
    var dst_len = len(dst)
    var res = alloc byte[dst_len + q_len]

    var i = 0
    while (i < dst_len) {
        res[i] = dst[i]
        i = i + 1
    }

    var j = 0
    while (j < q_len) {
        res[dst_len + j] = q[j]
        j = j + 1
    }
    return res
}
