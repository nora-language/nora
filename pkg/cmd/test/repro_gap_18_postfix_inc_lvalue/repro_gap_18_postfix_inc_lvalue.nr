// Gap 2.4c: ++ and -- in pkg/parser/parser.go:771 only parse on bare identifiers,
// failing to parse on struct fields (c.count++) or array elements (arr[0]++).
package main

import "os"

type Counter = struct {
    count: i32
}

fn main() i32 {
    var c = alloc Counter { count: 0 }
    c.count++
    if c.count != 1 {
        os.Exit(1)
    }
    return 0
}
