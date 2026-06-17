package main

import "io"
import "json"

fn test_primitives() {
    io.PrintLn("--- Test Primitives ---")

    // Test null
    var null_res = json.Parse("null")
    if (null_res.IsErr()) {
        panic("failed to parse null")
    }
    var null_val = null_res.Unwrap()
    if (!null_val.IsNull()) {
        panic("expected null")
    }
    io.PrintLn("Parsed null: ${json.Stringify(#null_val)}")

    // Test bools
    var true_res = json.Parse("true")
    var true_val = true_res.Unwrap()
    if (!true_val.GetBool()) {
        panic("expected true")
    }
    io.PrintLn("Parsed true: ${json.Stringify(#true_val)}")

    var false_res = json.Parse("false")
    if (false_res.IsErr()) {
        panic("failed to parse false")
    }
    var false_val = false_res.Unwrap()
    if (false_val.GetBool()) {
        panic("expected false")
    }
    io.PrintLn("Parsed false: ${json.Stringify(#false_val)}")

    // Test numbers
    var num_res = json.Parse("123.45")
    var num_val = num_res.Unwrap()
    var num = num_val.GetNumber()
    io.PrintLn("Parsed number: ${num}")
    if (num < 123.4 || num > 123.5) {
        panic("expected 123.45")
    }

    // Test string
    var str_res = json.Parse("\"Hello \\n\\t Nora!\"")
    if (str_res.IsErr()) {
        panic("failed to parse string with escape sequences")
    }
    var str_val = str_res.Unwrap()
    io.PrintLn("Parsed string: ${str_val.GetString()}")
}

fn test_array() {
    io.PrintLn("--- Test Array ---")
    var arr_res = json.Parse("[1.5, true, \"hello\", null]")
    if (arr_res.IsErr()) {
        panic("failed to parse array")
    }
    var arr_val = arr_res.Unwrap()
    var arr = arr_val.GetArray()

    if (arr.Len() != 4) {
        panic("expected length 4")
    }

    var item0 = arr.Get(0)
    var item1 = arr.Get(1)
    var item2 = arr.Get(2)
    var item3 = arr.Get(3)

    if (item0.GetNumber() < 1.4 || item0.GetNumber() > 1.6) {
        panic("expected index 0 to be 1.5")
    }
    if (!item1.GetBool()) {
        panic("expected index 1 to be true")
    }
    if (item2.GetString() != "hello") {
        panic("expected index 2 to be hello")
    }
    if (!item3.IsNull()) {
        panic("expected index 3 to be null")
    }

    io.PrintLn("Array stringified: ${json.Stringify(#arr_val)}")
}

fn test_object() {
    io.PrintLn("--- Test Object ---")
    var obj_res = json.Parse("{\"success\": true, \"value\": 42.0}")
    if (obj_res.IsErr()) {
        panic("failed to parse object")
    }
    var obj_val = obj_res.Unwrap()
    var obj = obj_val.GetObject()

    if (obj.Len() != 2) {
        panic("expected length 2")
    }

    var prop0 = obj.Get(0)
    var prop1 = obj.Get(1)

    io.PrintLn("Property 0: name=${prop0.name}, value=${json.Stringify(#prop0.value)}")
    io.PrintLn("Property 1: name=${prop1.name}, value=${json.Stringify(#prop1.value)}")

    if (prop0.name != "success" || !prop0.value.GetBool()) {
        panic("expected success property to be true")
    }
    if (prop1.name != "value" || prop1.value.GetNumber() != 42.0) {
        panic("expected value property to be 42.0")
    }

    io.PrintLn("Object stringified: ${json.Stringify(#obj_val)}")
}

fn test_errors() {
    io.PrintLn("--- Test Errors ---")
    var err_res1 = json.Parse("{\"invalid\"")
    if (err_res1.IsOk()) {
        panic("expected error on invalid JSON")
    }
    io.PrintLn("Gracefully caught invalid JSON: ${err_res1.UnwrapErr()}")

    var err_res2 = json.Parse("[1, 2,")
    if (err_res2.IsOk()) {
        panic("expected error on incomplete array")
    }
    io.PrintLn("Gracefully caught incomplete array: ${err_res2.UnwrapErr()}")
}

fn main() i32 {
    test_primitives()
    test_array()
    test_object()
    test_errors()

    io.PrintLn("--- All JSON standard library tests passed! ---")
    return 0
}
