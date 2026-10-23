// Gap 4.2b: Vector.Pop and Vector.Remove (std/collections/vector.nr:72-99) print an error message
// on out-of-bounds/empty access and fall through without panicking or returning, corrupting v.size to -1!
package main

import "collections"
import "os"

fn pop_empty(v: &collections.Vector[i32]) {
    var _ = v.Pop[i32]()
}

[unsafe]
fn main() i32 {
    var v = collections.NewVector[i32](4)
    var mon_ch = make(chan[str], 1)
    scope {
        spawn(mon_ch) pop_empty(&v)
    }
    var msg = <-mon_ch
    if msg != "Pop from empty vector" || v.Len[i32]() != 0 {
        os.Exit(1)
    }
    return 0
}
