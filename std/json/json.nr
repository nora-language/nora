package json

import "collections"
import "io"
import "string"

extern fn nr_atof(s: #str) f64

pub type JsonValue = enum {
    JsonNull,
    JsonBool(val: bool),
    JsonNumber(val: f64),
    JsonString(val: str),
    JsonArray(val: @collections.Vector[JsonValue]),
    JsonObject(val: @collections.Vector[JsonProperty])
}

pub type JsonProperty = struct {
    name: str,
    value: @JsonValue
}

type Parser = struct {
    input: #str,
    len: i32,
    pos: i32
}

pub fn Parse(s: #str) Result[JsonValue, str] {
    var p = Parser {
        input: s,
        len: string.Length(s),
        pos: 0
    }
    p.skipWhitespace()
    var val_res = p.parseValue()
    if (val_res.IsErr()) {
        return Err[JsonValue, str](val_res.UnwrapErr())
    }
    p.skipWhitespace()
    io.PrintLn("DEBUG Parse: pos=${p.pos}, len=${p.len}, input='${p.input}'")
    if (p.pos < p.len) {
        return Err[JsonValue, str]("Unexpected trailing characters after JSON")
    }
    return Ok[JsonValue, str](val_res.Unwrap())
}

pub fn Stringify(v: #JsonValue) str {
    match v {
        JsonNull => {
            return "null"
        }
        JsonBool(b) => {
            if (b) {
                return "true"
            }
            return "false"
        }
        JsonNumber(n) => {
            return io.FormatFloat(n)
        }
        JsonString(s) => {
            return escapeString(s)
        }
        JsonArray(arr) => {
            var parts = collections.NewVector[str](arr.Len())
            var i = 0
            while (i < arr.Len()) {
                var item = arr.Get(i)
                parts.Push(Stringify(#item))
                i = i + 1
            }
            var inner = string.Join(parts, ", ")
            return "[${inner}]"
        }
        JsonObject(obj) => {
            var parts = collections.NewVector[str](obj.Len())
            var i = 0
            while (i < obj.Len()) {
                var prop = obj.Get(i)
                var escaped_key = escapeString(#prop.name)
                var val_str = Stringify(#prop.value)
                parts.Push("${escaped_key}: ${val_str}")
                i = i + 1
            }
            var inner = string.Join(parts, ", ")
            return "{${inner}}"
        }
    }
}

fn escapeString(s: #str) str {
    var len_s = string.Length(s)
    var max_len = (len_s * 2) + 3
    var buf = alloc byte[max_len]
    buf[0] = 34
    var buf_idx = 1

    var i = 0
    while (i < len_s) {
        var c = s[i]
        if (c == 34) {
            buf[buf_idx] = 92
            buf[buf_idx + 1] = 34
            buf_idx = buf_idx + 2
        } else if (c == 92) {
            buf[buf_idx] = 92
            buf[buf_idx + 1] = 92
            buf_idx = buf_idx + 2
        } else if (c == 10) {
            buf[buf_idx] = 92
            buf[buf_idx + 1] = 110
            buf_idx = buf_idx + 2
        } else if (c == 9) {
            buf[buf_idx] = 92
            buf[buf_idx + 1] = 116
            buf_idx = buf_idx + 2
        } else if (c == 13) {
            buf[buf_idx] = 92
            buf[buf_idx + 1] = 114
            buf_idx = buf_idx + 2
        } else {
            buf[buf_idx] = c
            buf_idx = buf_idx + 1
        }
        i = i + 1
    }
    buf[buf_idx] = 34
    buf[buf_idx + 1] = 0
    return buf
}

fn (self: &Parser) skipWhitespace() {
    while (self.pos < self.len) {
        var c = self.input[self.pos]
        if (c == 32 || c == 9 || c == 10 || c == 13) {
            self.pos = self.pos + 1
        } else {
            break;
        }
    }
}

fn (self: &Parser) parseValue() Result[JsonValue, str] {
    self.skipWhitespace()
    if (self.pos >= self.len) {
        return Err[JsonValue, str]("Unexpected end of input")
    }
    var c = self.input[self.pos]
    if (c == 123) {
        return self.parseObject()
    }
    if (c == 91) {
        return self.parseArray()
    }
    if (c == 34) {
        var str_res = self.parseString()
        if (str_res.IsErr()) {
            return Err[JsonValue, str](str_res.UnwrapErr())
        }
        return Ok[JsonValue, str](JsonString(str_res.Unwrap()))
    }
    if (c == 110) {
        return self.parseNull()
    }
    if (c == 116 || c == 102) {
        return self.parseBool()
    }
    if ((c >= 48 && c <= 57) || c == 45) {
        return self.parseNumber()
    }
    return Err[JsonValue, str]("Unexpected character in JSON")
}

fn (self: &Parser) parseNull() Result[JsonValue, str] {
    if (self.pos + 4 <= self.len && self.input[self.pos] == 110 &&
        self.input[self.pos + 1] == 117 && self.input[self.pos + 2] == 108 &&
        self.input[self.pos + 3] == 108) {
        self.pos = self.pos + 4
        return Ok[JsonValue, str](JsonNull)
    }
    return Err[JsonValue, str]("Expected 'null'")
}

fn (self: &Parser) parseBool() Result[JsonValue, str] {
    if (self.pos + 4 <= self.len && self.input[self.pos] == 116 &&
        self.input[self.pos + 1] == 114 && self.input[self.pos + 2] == 117 &&
        self.input[self.pos + 3] == 101) {
        self.pos = self.pos + 4
        return Ok[JsonValue, str](JsonBool(true))
    }
    if (self.pos + 5 <= self.len && self.input[self.pos] == 102 &&
        self.input[self.pos + 1] == 97 && self.input[self.pos + 2] == 108 &&
        self.input[self.pos + 3] == 115 && self.input[self.pos + 4] == 101) {
        self.pos = self.pos + 5
        return Ok[JsonValue, str](JsonBool(false))
    }
    return Err[JsonValue, str]("Expected 'true' or 'false'")
}

fn (self: &Parser) parseNumber() Result[JsonValue, str] {
    var start = self.pos
    while (self.pos < self.len) {
        var c = self.input[self.pos]
        if ((c >= 48 && c <= 57) || c == 46 || c == 45 ||
            c == 43 || c == 101 || c == 69) {
            self.pos = self.pos + 1
        } else {
            break;
        }
    }
    if (self.pos == start) {
        return Err[JsonValue, str]("Expected number")
    }
    var num_str = self.substring(start, self.pos)
    var val = nr_atof(num_str)
    return Ok[JsonValue, str](JsonNumber(val))
}

fn (self: #Parser) substring(start: i32, end: i32) str {
    var len = end - start
    if (len <= 0) {
        return ""
    }
    var b = alloc byte[len + 1]
    var i = 0
    while (i < len) {
        b[i] = self.input[start + i]
        i = i + 1
    }
    b[len] = 0
    return b
}

fn (self: &Parser) parseString() Result[str, str] {
    if (self.pos >= self.len || self.input[self.pos] != 34) {
        return Err[str, str]("Expected '\"' at start of string")
    }
    self.pos = self.pos + 1

    var max_len = self.len - self.pos
    var buf = alloc byte[max_len + 1]
    var buf_idx = 0

    while (self.pos < self.len) {
        var c = self.input[self.pos]
        if (c == 34) {
            self.pos = self.pos + 1
            buf[buf_idx] = 0
            return Ok[str, str](buf)
        }
        if (c == 92) {
            if (self.pos + 1 >= self.len) {
                return Err[str, str]("Unterminated string escape")
            }
            var next_c = self.input[self.pos + 1]
            self.pos = self.pos + 2
            if (next_c == 34) {
                buf[buf_idx] = 34
            } else if (next_c == 92) {
                buf[buf_idx] = 92
            } else if (next_c == 47) {
                buf[buf_idx] = 47
            } else if (next_c == 98) {
                buf[buf_idx] = 8
            } else if (next_c == 102) {
                buf[buf_idx] = 12
            } else if (next_c == 110) {
                buf[buf_idx] = 10
            } else if (next_c == 114) {
                buf[buf_idx] = 13
            } else if (next_c == 116) {
                buf[buf_idx] = 9
            } else {
                return Err[str, str]("Unsupported string escape sequence")
            }
            buf_idx = buf_idx + 1
        } else {
            buf[buf_idx] = c
            buf_idx = buf_idx + 1
            self.pos = self.pos + 1
        }
    }
    return Err[str, str]("Unterminated string in JSON")
}

fn (self: &Parser) parseArray() Result[JsonValue, str] {
    if (self.pos >= self.len || self.input[self.pos] != 91) {
        return Err[JsonValue, str]("Expected '[' at start of array")
    }
    self.pos = self.pos + 1
    self.skipWhitespace()

    var vec = collections.NewVector[JsonValue](4)
    if (self.pos < self.len && self.input[self.pos] == 93) {
        self.pos = self.pos + 1
        return Ok[JsonValue, str](JsonArray(vec))
    }

    while (self.pos < self.len) {
        self.skipWhitespace()
        var val_res = self.parseValue()
        if (val_res.IsErr()) {
            return Err[JsonValue, str](val_res.UnwrapErr())
        }
        vec.Push(val_res.Unwrap())

        self.skipWhitespace()
        if (self.pos >= self.len) {
            return Err[JsonValue, str]("Unexpected end of input in array")
        }
        var next_c = self.input[self.pos]
        if (next_c == 93) {
            self.pos = self.pos + 1
            return Ok[JsonValue, str](JsonArray(vec))
        }
        if (next_c == 44) {
            self.pos = self.pos + 1
        } else {
            return Err[JsonValue, str]("Expected ',' or ']' in array")
        }
    }
    return Err[JsonValue, str]("Unterminated array in JSON")
}

fn (self: &Parser) parseObject() Result[JsonValue, str] {
    if (self.pos >= self.len || self.input[self.pos] != 123) {
        return Err[JsonValue, str]("Expected '{' at start of object")
    }
    self.pos = self.pos + 1
    self.skipWhitespace()

    var vec = collections.NewVector[JsonProperty](4)
    if (self.pos < self.len && self.input[self.pos] == 125) {
        self.pos = self.pos + 1
        return Ok[JsonValue, str](JsonObject(vec))
    }

    while (self.pos < self.len) {
        self.skipWhitespace()
        if (self.pos >= self.len || self.input[self.pos] != 34) {
            return Err[JsonValue, str]("Expected string key in object")
        }
        var key_res = self.parseString()
        if (key_res.IsErr()) {
            return Err[JsonValue, str](key_res.UnwrapErr())
        }
        var key = key_res.Unwrap()

        self.skipWhitespace()
        if (self.pos >= self.len || self.input[self.pos] != 58) {
            return Err[JsonValue, str]("Expected ':' after object key")
        }
        self.pos = self.pos + 1

        self.skipWhitespace()
        var val_res = self.parseValue()
        if (val_res.IsErr()) {
            return Err[JsonValue, str](val_res.UnwrapErr())
        }
        var val = val_res.Unwrap()

        var val_ptr = alloc @val
        var prop = JsonProperty {
            name: key,
            value: @val_ptr
        }
        vec.Push(prop)

        self.skipWhitespace()
        if (self.pos >= self.len) {
            return Err[JsonValue, str]("Unexpected end of input in object")
        }
        var next_c = self.input[self.pos]
        if (next_c == 125) {
            self.pos = self.pos + 1
            return Ok[JsonValue, str](JsonObject(vec))
        }
        if (next_c == 44) {
            self.pos = self.pos + 1
        } else {
            return Err[JsonValue, str]("Expected ',' or '}' in object")
        }
    }
    return Err[JsonValue, str]("Unterminated object in JSON")
}

pub fn (self: #JsonValue) IsNull() bool {
    match self {
        JsonNull => {
            return true
        }
        _ => {
            return false
        }
    }
}

pub fn (self: #JsonValue) GetBool() bool {
    match self {
        JsonBool(b) => {
            return b
        }
        _ => {
            panic("called GetBool() on non-bool JsonValue")
        }
    }
}

pub fn (self: #JsonValue) GetNumber() f64 {
    match self {
        JsonNumber(n) => {
            return n
        }
        _ => {
            panic("called GetNumber() on non-number JsonValue")
        }
    }
}

pub fn (self: #JsonValue) GetString() #str {
    match self {
        JsonString(s) => {
            return s
        }
        _ => {
            panic("called GetString() on non-string JsonValue")
        }
    }
}

pub fn (self: #JsonValue) GetArray() #collections.Vector[JsonValue] {
    match self {
        JsonArray(arr) => {
            return arr
        }
        _ => {
            panic("called GetArray() on non-array JsonValue")
        }
    }
}

pub fn (self: #JsonValue) GetObject() #collections.Vector[JsonProperty] {
    match self {
        JsonObject(obj) => {
            return obj
        }
        _ => {
            panic("called GetObject() on non-object JsonValue")
        }
    }
}
