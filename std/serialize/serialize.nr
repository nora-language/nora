package serialize

import "io"
import "json"
import "strconv"
import "string"

pub fn ToJSON[T](val: #T) str {
    return nr_serialize_json[T](val)
}

pub fn FromJSON[T](data: str) T {
    return nr_deserialize_json[T](data)
}

// Built-in extern intrinsics
[plugin_override]
extern fn nr_serialize_json[T](val: #T) str

[plugin_override]
extern fn nr_deserialize_json[T](data: str) T

// --- Primitive Serialization Helpers ---

pub fn nr_serialize_json_i32(val: #i32) str {
    return io.FormatInt(val)
}

pub fn nr_deserialize_json_i32(data: str) @i32 {
    var v_res = json.Parse(data)
    if (v_res.IsErr()) { return 0 }
    var v = v_res.Unwrap()
    return i32(v.GetNumber())
}

pub fn nr_serialize_json_i64(val: #i64) str {
    return io.FormatInt(i32(val)) // Format via standard i32 for simplicity or strconv
}

pub fn nr_deserialize_json_i64(data: str) @i64 {
    var v_res = json.Parse(data)
    if (v_res.IsErr()) { return i64(0) }
    var v = v_res.Unwrap()
    return i64(v.GetNumber())
}

pub fn nr_serialize_json_f64(val: #f64) str {
    return io.FormatFloat(val)
}

pub fn nr_deserialize_json_f64(data: str) @f64 {
    var v_res = json.Parse(data)
    if (v_res.IsErr()) { return 0.0 }
    var v = v_res.Unwrap()
    return v.GetNumber()
}

pub fn nr_serialize_json_f32(val: #f32) str {
    return io.FormatFloat(f64(val))
}

pub fn nr_deserialize_json_f32(data: str) @f32 {
    var v_res = json.Parse(data)
    if (v_res.IsErr()) { return f32(0.0) }
    var v = v_res.Unwrap()
    return f32(v.GetNumber())
}

pub fn nr_serialize_json_bool(val: #bool) str {
    if (val == true) {
        return "true"
    }
    return "false"
}

pub fn nr_deserialize_json_bool(data: str) @bool {
    var v_res = json.Parse(data)
    if (v_res.IsErr()) { return false }
    var v = v_res.Unwrap()
    return v.GetBool()
}

pub fn nr_serialize_json_str(val: #str) str {
    return EscapeString(val)
}

pub fn nr_deserialize_json_str(data: str) @str {
    var v_res = json.Parse(data)
    if (v_res.IsErr()) { return "" }
    var v = v_res.Unwrap()
    return "" + v.GetString()
}

// Helper to escape strings in JSON format
pub fn EscapeString(s: #str) str {
    var len_s = string.Length(s)
    var max_len = (len_s * 2) + 3
    var buf = alloc byte[max_len]
    buf[0] = 34 // '"'
    var buf_idx = 1

    var i = 0
    while (i < len_s) {
        var c = s[i]
        if (c == 34) {
            buf[buf_idx] = 92 // '\'
            buf[buf_idx + 1] = 34 // '"'
            buf_idx = buf_idx + 2
        } else if (c == 92) {
            buf[buf_idx] = 92
            buf[buf_idx + 1] = 92
            buf_idx = buf_idx + 2
        } else if (c == 10) { // '\n'
            buf[buf_idx] = 92
            buf[buf_idx + 1] = 110
            buf_idx = buf_idx + 2
        } else if (c == 9) { // '\t'
            buf[buf_idx] = 92
            buf[buf_idx + 1] = 116
            buf_idx = buf_idx + 2
        } else if (c == 13) { // '\r'
            buf[buf_idx] = 92
            buf[buf_idx + 1] = 114
            buf_idx = buf_idx + 2
        } else {
            buf[buf_idx] = c
            buf_idx = buf_idx + 1
        }
        i = i + 1
    }
    buf[buf_idx] = 34 // '"'
    buf[buf_idx + 1] = 0
    return buf
}
