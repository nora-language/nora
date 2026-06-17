package string

extern fn strlen(s: str) i32
extern fn strcmp(s1: str, s2: str) i32
extern fn strncmp(s1: str, s2: str, n: i32) i32
extern fn strstr(haystack: str, needle: str) str
extern fn toupper(c: i32) i32
extern fn tolower(c: i32) i32
extern fn isspace(c: i32) i32

fn length(s: str) i32 {
    return strlen(s)
}

fn equals(s1: str, s2: str) bool {
    return strcmp(s1, s2) == 0
}

fn contains(s: str, substr: str) bool {
    var res = strstr(s, substr)
    return res != none
}

fn has_prefix(s: str, prefix: str) bool {
    var lp = length(prefix)
    var ls = length(s)
    if (lp > ls) {
        return false
    }
    return strncmp(s, prefix, lp) == 0
}

fn has_suffix(s: str, suffix: str) bool {
    var ls = length(s)
    var lsu = length(suffix)
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

fn to_upper(s: str) str {
    var n = length(s)
    var b = alloc byte[n + 1]
    var i = 0
    while i < n {
        b[i] = toupper(s[i])
        i = i + 1
    }
    b[n] = 0
    return b
}

fn to_lower(s: str) str {
    var n = length(s)
    var b = alloc byte[n + 1]
    var i = 0
    while i < n {
        b[i] = tolower(s[i])
        i = i + 1
    }
    b[n] = 0
    return b
}

fn trim(s: str) str {
    var n = length(s)
    if (n == 0) {
        return ""
    }
    var start = 0
    while start < n && isspace(s[start]) != 0 {
        start = start + 1
    }
    if (start == n) {
        return ""
    }
    var end = n - 1
    while end > start && isspace(s[end]) != 0 {
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
