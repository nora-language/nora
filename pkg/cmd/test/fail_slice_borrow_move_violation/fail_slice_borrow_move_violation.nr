package main

import "io"

fn main() i32 {
    var arr = alloc str[5]
    arr[0] = "alpha"
    arr[1] = "beta"

    // Sub-slice leases 'arr'
    var sub = arr[0..2]

    // ERROR: Topological solver must reject moving 'arr' while 'sub' is actively leased!
    var moved = @arr

    io.PrintLn("Sub zero:", sub[0])
    return 0
}
