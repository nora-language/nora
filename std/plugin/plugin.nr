package plugin

extern fn strlen(s: str) i32

extern fn strstr(haystack: str, needle: str) ptr

pub export fn get_json_value(payload: str, key: str) str {
    var key_with_quotes = "\"${key}\":"
    var found: str = strstr(payload, key_with_quotes)
    if (found == none) {
        return ""
    }

    var offset = strlen(key_with_quotes)
    while (found[offset] == 32) { // space
        offset = offset + 1
    }

    if (found[offset] != 34) { // quote
        return ""
    }
    offset = offset + 1
    var start = offset

    while (found[offset] != 34 && found[offset] != 0) {
        offset = offset + 1
    }

    var length = offset - start
    if (length <= 0) {
        return ""
    }

    var res = alloc i8[length + 1]
    var i = 0
    while (i < length) {
        res[i] = found[start + i]
        i = i + 1
    }
    res[length] = 0

    return res
}

pub export fn get_json_array_element(payload: str, key: str, index: i32) str {
    var key_with_quotes = "\"${key}\":"
    var found: str = strstr(payload, key_with_quotes)
    if (found == none) {
        return ""
    }

    var offset = strlen(key_with_quotes)
    while (found[offset] != 91 && found[offset] != 0) { // '['
        offset = offset + 1
    }
    if (found[offset] == 0) {
        return ""
    }
    offset = offset + 1 // skip [

    var curr = 0
    while (curr < index) {
        var depth = 0
        var in_str = false
        var skip_done = false
        while (!skip_done) {
            if (!in_str) {
                if (found[offset] == 123 || found[offset] == 91) {
                    depth = depth + 1
                }
                if (found[offset] == 125 || found[offset] == 93) {
                    depth = depth - 1
                }
            }
            if (found[offset] == 34) {
                in_str = !in_str
            }

            if (depth == 0 && !in_str && found[offset] == 44) {
                offset = offset + 1
                curr = curr + 1
                skip_done = true
            }
            if (depth < 0 || found[offset] == 0) {
                return ""
            }
            if (!skip_done) {
                offset = offset + 1
            }
        }
    }

    while (found[offset] == 32) {
        offset = offset + 1
    }
    var start = offset

    var depth = 0
    var in_str = false
    var element_done = false
    while (!element_done) {
        if (!in_str) {
            if (found[offset] == 123 || found[offset] == 91) {
                depth = depth + 1
            }
            if (found[offset] == 125 || found[offset] == 93) {
                depth = depth - 1
            }
        }
        if (found[offset] == 34) {
            in_str = !in_str
        }

        offset = offset + 1
        if (depth == 0 && !in_str && (found[offset] == 44 ||
            found[offset] == 93)) {
            element_done = true
        }
        if (found[offset] == 0) {
            element_done = true
        }
    }

    var length = offset - start
    if (length <= 0) {
        return ""
    }

    var res = alloc i8[length + 1]
    var i = 0
    while (i < length) {
        res[i] = found[start + i]
        i = i + 1
    }
    res[length] = 0
    return res
}
