package main

import "std/collections"
import "io"
import "strconv"

type BodyInfo = struct {
    x: i32,
    y: i32
}

fn main() i32 {
    var vec = collections.NewVector[BodyInfo](4)
    vec.Push(BodyInfo { x: 10, y: 20 })
    vec.Push(BodyInfo { x: 30, y: 40 })
    vec.Push(BodyInfo { x: 50, y: 60 })
    vec.Push(BodyInfo { x: 70, y: 80 })

    io.PrintLn("Before ParMap:")
    var i = 0
    while i < vec.Len[BodyInfo]() {
        var el = #vec.Get[BodyInfo](i)
        io.PrintLn(strconv.FormatInt(el.x, 10))
        i = i + 1
    }

    vec.ParMap[BodyInfo](fn(b: &BodyInfo) {
        // Multiply x by 2 in parallel
        b.x = b.x * 2
    })

    io.PrintLn("After ParMap:")
    i = 0
    while i < vec.Len[BodyInfo]() {
        var el = #vec.Get[BodyInfo](i)
        io.PrintLn(strconv.FormatInt(el.x, 10))
        i = i + 1
    }
    
    // Verify results
    if vec.Get[BodyInfo](0).x != 20 { return 1 }
    if vec.Get[BodyInfo](1).x != 60 { return 1 }
    if vec.Get[BodyInfo](2).x != 100 { return 1 }
    if vec.Get[BodyInfo](3).x != 140 { return 1 }

    io.PrintLn("ParMap completed successfully!")

    return 0
}
