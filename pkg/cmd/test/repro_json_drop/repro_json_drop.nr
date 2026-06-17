package main

import "io"
import "json"

fn test_repro() {
    // Parse nested object which produces nested JsonValue (JsonObject containing JsonProperty containing JsonValue)
    var res = json.Parse("{\"inner\": {\"a\": 1, \"b\": 2}}")
    if (res.IsErr()) {
        panic("Parse failed")
    }
    var val = res.Unwrap()
    
    // Stringify to trigger stringify logic
    var s = json.Stringify(#val)
    io.PrintLn("Stringified: ${s}")
    
    // Explicit drop occurs at the end of this scope when `val` is dropped.
    // If drop logic performs `nr_free` on inline union field, it will segfault here.
}

fn main() i32 {
    io.PrintLn("Starting repro_json_drop test...")
    test_repro()
    io.PrintLn("Success!")
    return 0
}
