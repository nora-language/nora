package json

import "string"
import "ffi"

pub fn GetProp(val: #JsonValue, key: #str) ptr {
    match val {
        JsonObject(obj) => {
            var i = 0
            while (i < obj.Len[JsonProperty]()) {
                var prop = obj.Get[JsonProperty](i)
                if (string.Compare(#prop.name, key) == 0) {
                    return ffi.BorrowToRaw[JsonValue](#prop.value)
                }
                i = i + 1
            }
        }
        _ => {}
    }
    return ptr(0)
}

pub fn GetArrayItem(val: #JsonValue, index: i32) ptr {
    match val {
        JsonArray(arr) => {
            if (index >= 0 && index < arr.Len[JsonValue]()) {
                var item = arr.Get[JsonValue](index)
                return ffi.BorrowToRaw[JsonValue](#item)
            }
        }
        _ => {}
    }
    return ptr(0)
}

pub fn GetNumber(val: #JsonValue) f64 {
    match val {
        JsonNumber(n) => { return n }
        _ => {}
    }
    return 0.0
}
