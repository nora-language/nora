// Gap 2.3: sa.implements (pkg/semantic/analyzer.go:7440) ignores ReceiverLease and ParamLeases,
// allowing a struct with a mutable &Self method to satisfy an interface and be called via a
// read-only #Self borrow, mutating a read-only borrowed struct!
package main

import "io"
import "os"

type Reader = interface {
    fn Read() i32
}

type Counter = struct {
    n: i32
}

// Requires mutable &Counter receiver and mutates self.n!
fn (self: &Counter) Read() i32 {
    self.n = self.n + 1
    return self.n
}

fn read_from_immutable(c: #Counter) i32 {
    // Coercing read-only borrow #Counter into Reader interface and calling Read()
    // mutates c.n even though c is a read-only (#) borrow!
    var r: Reader = c
    return r.Read()
}

fn main() i32 {
    var c = alloc Counter { n: 10 }
    var v = read_from_immutable(#c)
    if c.n != 10 {
        io.PrintLn("FAIL: read-only borrow #Counter was mutated via interface dispatch (c.n = ${c.n})")
        os.Exit(1)
    }
    return 0
}
