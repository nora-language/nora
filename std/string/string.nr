package string

import "collections"

extern fn strlen(s: #str) i32

extern fn strcmp(s1: #str, s2: #str) i32

extern fn strncmp(s1: #str, s2: #str, n: i32) i32

pub fn Clone(s: #str) str {
    return Substring(s, 0, Length(s))
}
extern fn strstr(haystack: #str, needle: #str) ptr

fn toupper(c: i32) i32 {
    if (c >= 97 && c <= 122) {
        return c - 32
    }
    return c
}

fn tolower(c: i32) i32 {
    if (c >= 65 && c <= 90) {
        return c + 32
    }
    return c
}

fn isspace(c: i32) bool {
    return c == 32 || (c >= 9 && c <= 13)
}

fn isdigit(c: i32) bool {
    return c >= 48 && c <= 57
}

fn isalpha(c: i32) bool {
    return (c >= 65 && c <= 90) || (c >= 97 && c <= 122)
}

fn isalnum(c: i32) bool {
    return (c >= 48 && c <= 57) || (c >= 65 && c <= 90) ||
        (c >= 97 && c <= 122)
}

// ─── Metadata & Comparison ───────────────────────────────────────────────────
pub fn Length(s: #str) i32 {
    return strlen(s)
}

pub fn Equals(s1: #str, s2: #str) bool {
    return strcmp(s1, s2) == 0
}

pub fn Compare(s1: #str, s2: #str) i32 {
    return strcmp(s1, s2)
}

// ─── Substrings & Slicing ────────────────────────────────────────────────────
pub fn Contains(s: #str, substr: #str) bool {
    var res: ptr = strstr(s, substr)
    return res != none
}

pub fn HasPrefix(s: #str, prefix: #str) bool {
    var lp = strlen(prefix)
    var ls = strlen(s)
    if (lp > ls) {
        return false
    }
    return strncmp(s, prefix, lp) == 0
}

pub fn HasSuffix(s: #str, suffix: #str) bool {
    var ls = strlen(s)
    var lsu = strlen(suffix)
    if (lsu > ls) {
        return false
    }
    var offset = ls - lsu
    var i = 0
    while i < lsu {
        if (s[offset + i] != suffix[i]) {
            return false
        }
        i = i + 1
    }
    return true
}

[unsafe]
pub fn Substring(s: #str, start: i32, end: i32) str {
    var len = strlen(s)
    var s_idx = start
    var e_idx = end

    if (s_idx < 0) {
        s_idx = len + s_idx
        if (s_idx < 0) {
            s_idx = 0
        }
    }
    if (s_idx > len) {
        s_idx = len
    }

    if (e_idx < 0) {
        e_idx = len + e_idx
        if (e_idx < 0) {
            e_idx = 0
        }
    }
    if (e_idx > len) {
        e_idx = len
    }

    if (s_idx >= e_idx) {
        return ""
    }

    var sub_len = e_idx - s_idx
    var b = alloc byte[sub_len + 1]
    var i = 0
    while i < sub_len {
        b[i] = s[s_idx + i]
        i = i + 1
    }
    b[sub_len] = 0
    return b
}

// ─── Search & Indexing ────────────────────────────────────────────────────────
pub fn IndexOf(s: #str, substr: #str) i32 {
    var ls = strlen(s)
    var lsub = strlen(substr)
    if (lsub == 0) {
        return 0
    }
    if (lsub > ls) {
        return -1
    }
    var limit = ls - lsub
    var i = 0
    while i <= limit {
        var match_found = true
        var j = 0
        while j < lsub {
            if (s[i + j] != substr[j]) {
                match_found = false
                break;
            }
            j = j + 1
        }
        if (match_found) {
            return i
        }
        i = i + 1
    }
    return -1
}

pub fn LastIndexOf(s: #str, substr: #str) i32 {
    var ls = strlen(s)
    var lsub = strlen(substr)
    if (lsub == 0) {
        return ls
    }
    if (lsub > ls) {
        return -1
    }
    var limit = ls - lsub
    var i = limit
    while i >= 0 {
        var match_found = true
        var j = 0
        while j < lsub {
            if (s[i + j] != substr[j]) {
                match_found = false
                break;
            }
            j = j + 1
        }
        if (match_found) {
            return i
        }
        i = i - 1
    }
    return -1
}

// ─── Case Conversion & Whitespace ────────────────────────────────────────────
[unsafe]
pub fn ToUpper(s: #str) str {
    var n = strlen(s)
    var b = alloc byte[n + 1]
    var i = 0
    while i < n {
        b[i] = toupper(s[i])
        i = i + 1
    }
    b[n] = 0
    return b
}

[unsafe]
pub fn ToLower(s: #str) str {
    var n = strlen(s)
    var b = alloc byte[n + 1]
    var i = 0
    while i < n {
        b[i] = tolower(s[i])
        i = i + 1
    }
    b[n] = 0
    return b
}

[unsafe]
pub fn Trim(s: #str) str {
    var n = strlen(s)
    if (n == 0) {
        return ""
    }
    var start = 0
    while start < n && isspace(s[start]) {
        start = start + 1
    }
    if (start == n) {
        return ""
    }
    var end = n - 1
    while end > start && isspace(s[end]) {
        end = end - 1
    }
    var new_n = end - start + 1
    var b = alloc byte[new_n + 1]
    var i = 0
    while i < new_n {
        b[i] = s[start + i]
        i = i + 1
    }
    b[new_n] = 0
    return b
}

[unsafe]
pub fn TrimLeft(s: #str) str {
    var n = strlen(s)
    if (n == 0) {
        return ""
    }
    var start = 0
    while start < n && isspace(s[start]) {
        start = start + 1
    }
    if (start == n) {
        return ""
    }
    var new_n = n - start
    var b = alloc byte[new_n + 1]
    var i = 0
    while i < new_n {
        b[i] = s[start + i]
        i = i + 1
    }
    b[new_n] = 0
    return b
}

[unsafe]
pub fn TrimRight(s: #str) str {
    var n = strlen(s)
    if (n == 0) {
        return ""
    }
    var end = n - 1
    while end >= 0 && isspace(s[end]) {
        end = end - 1
    }
    if (end < 0) {
        return ""
    }
    var new_n = end + 1
    var b = alloc byte[new_n + 1]
    var i = 0
    while i < new_n {
        b[i] = s[i]
        i = i + 1
    }
    b[new_n] = 0
    return b
}

// ─── Characters & Classifications ────────────────────────────────────────────
pub fn IsNumeric(s: #str) bool {
    var n = strlen(s)
    if (n == 0) {
        return false
    }
    var i = 0
    while i < n {
        if (!isdigit(s[i])) {
            return false
        }
        i = i + 1
    }
    return true
}

pub fn IsAlpha(s: #str) bool {
    var n = strlen(s)
    if (n == 0) {
        return false
    }
    var i = 0
    while i < n {
        if (!isalpha(s[i])) {
            return false
        }
        i = i + 1
    }
    return true
}

pub fn IsAlphaNumeric(s: #str) bool {
    var n = strlen(s)
    if (n == 0) {
        return false
    }
    var i = 0
    while i < n {
        if (!isalnum(s[i])) {
            return false
        }
        i = i + 1
    }
    return true
}

// ─── Transformations ─────────────────────────────────────────────────────────
[unsafe]
pub fn Repeat(s: #str, count: i32) str {
    if (count <= 0) {
        return ""
    }
    var len = strlen(s)
    if (len == 0) {
        return ""
    }
    var total_len = len * count
    var b = alloc byte[total_len + 1]
    var i = 0
    while i < count {
        var j = 0
        while j < len {
            b[i * len + j] = s[j]
            j = j + 1
        }
        i = i + 1
    }
    b[total_len] = 0
    return b
}

[unsafe]
pub fn Reverse(s: #str) str {
    var len = strlen(s)
    if (len == 0) {
        return ""
    }
    var b = alloc byte[len + 1]
    var i = 0
    while i < len {
        b[i] = s[len - 1 - i]
        i = i + 1
    }
    b[len] = 0
    return b
}

[unsafe]
pub fn Replace(s: #str, old_str: #str, new_str: #str) str {
    var len_s = strlen(s)
    var len_old = strlen(old_str)
    var len_new = strlen(new_str)

    if (len_s == 0 || len_old == 0) {
        return "" + s
    }

    // Count occurrences
    var count = 0
    var i = 0
    while i <= len_s - len_old {
        var match_found = true
        var j = 0
        while j < len_old {
            if (s[i + j] != old_str[j]) {
                match_found = false
                break;
            }
            j = j + 1
        }
        if (match_found) {
            count = count + 1
            i = i + len_old
        } else {
            i = i + 1
        }
    }

    if (count == 0) {
        return "" + s
    }

    var total_len = len_s - (count * len_old) + (count * len_new)
    var b = alloc byte[total_len + 1]

    var src_idx = 0
    var dst_idx = 0
    while src_idx < len_s {
        var match_found = false
        if (src_idx <= len_s - len_old) {
            match_found = true
            var j = 0
            while j < len_old {
                if (s[src_idx + j] != old_str[j]) {
                    match_found = false
                    break;
                }
                j = j + 1
            }
        }

        if (match_found) {
            var k = 0
            while k < len_new {
                b[dst_idx + k] = new_str[k]
                k = k + 1
            }
            dst_idx = dst_idx + len_new
            src_idx = src_idx + len_old
        } else {
            b[dst_idx] = s[src_idx]
            dst_idx = dst_idx + 1
            src_idx = src_idx + 1
        }
    }
    b[total_len] = 0
    return b
}

pub fn FromChar(c: byte) str {
    var b = alloc byte[2]
    b[0] = c
    b[1] = 0
    return b
}

pub fn Split(s: #str, sep: #str) @collections.Vector[str] {
    var len_s = strlen(s)
    var len_sep = strlen(sep)
    var vec = collections.NewVector[str](4)

    if (len_s == 0) {
        vec.Push("")
        return vec
    }

    if (len_sep == 0) {
        // Split into individual characters as strings
        var i = 0
        while i < len_s {
            vec.Push(FromChar(s[i]))
            i = i + 1
        }
        return vec
    }

    var start = 0
    var i = 0
    while i <= len_s - len_sep {
        var match_found = true
        var j = 0
        while j < len_sep {
            if (s[i + j] != sep[j]) {
                match_found = false
                break;
            }
            j = j + 1
        }

        if (match_found) {
            vec.Push(Substring(s, start, i))
            start = i + len_sep
            i = start
        } else {
            i = i + 1
        }
    }

    // Push the remaining part
    vec.Push(Substring(s, start, len_s))

    return vec
}

[unsafe]
pub fn Join(parts: #collections.Vector[str], sep: #str) str {
    var num_parts = parts.Len()
    if (num_parts == 0) {
        return ""
    }

    var len_sep = strlen(sep)
    var total_len = 0

    var i = 0
    while i < num_parts {
        var part = parts.Get(i)
        total_len = total_len + strlen(part)
        if (i < num_parts - 1) {
            total_len = total_len + len_sep
        }
        i = i + 1
    }

    var b = alloc byte[total_len + 1]
    var dst_idx = 0

    i = 0
    while i < num_parts {
        var part = parts.Get(i)
        var len_part = strlen(part)

        var j = 0
        while j < len_part {
            b[dst_idx + j] = part[j]
            j = j + 1
        }
        dst_idx = dst_idx + len_part

        if (i < num_parts - 1) {
            var k = 0
            while k < len_sep {
                b[dst_idx + k] = sep[k]
                k = k + 1
            }
            dst_idx = dst_idx + len_sep
        }

        i = i + 1
    }
    b[total_len] = 0
    return b
}
