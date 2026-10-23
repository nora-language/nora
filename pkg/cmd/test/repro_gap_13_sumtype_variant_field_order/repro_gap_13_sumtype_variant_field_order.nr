// Gap 2.2a: analyzePattern (pkg/semantic/analyzer.go:1110) and selector constructors (line 4437)
// use sa.sortedFieldNames(variant.Fields) (alphabetical order) instead of variant.FieldNames (declaration order)!
// When a variant's field names are not alphabetical (e.g. z_code before a_label), types are swapped!
package main

import "io"
import "os"

type Event = enum {
    Tagged(z_code: i32, a_label: str)
    None
}

fn main() i32 {
    var ev = Event.Tagged(42, "ok")
    match ev {
        Tagged(code, label) => {
            if code != 42 {
                io.PrintLn("FAIL: expected code 42")
                os.Exit(1)
            }
            if label != "ok" {
                io.PrintLn("FAIL: expected label ok")
                os.Exit(1)
            }
        }
        None => {
            os.Exit(1)
        }
    }
    return 0
}
