// Gap 4.2b: Vector.Pop and Vector.Remove (std/collections/vector.nr:72-99) print an error message
// on out-of-bounds/empty access and fall through without panicking or returning, corrupting v.size to -1!
package main

import "collections"
import "io"
import "os"

fn main() i32 {
    var v = collections.NewVector[i32](4)
    var dummy = v.Pop[i32]()
    if v.Len[i32]() < 0 {
        io.PrintLn("FAIL: Vector.Pop on empty vector corrupted v.size to ${v.Len[i32]()}")
        os.Exit(1)
    }
    return 0
}
