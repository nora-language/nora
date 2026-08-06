package main

import "io"
import "os"

pub type Level3 = struct {
    payload: @i32[],
    tag: i32
}

pub type Level2 = struct {
    child: Level3,
    id: i32
}

pub type Level1 = struct {
    sub: Level2,
    name: i32
}

fn main() i32 {
    var arr = alloc i32[5]
    var root = Level1 {
        sub: Level2 {
            child: Level3 {
                payload: @arr,
                tag: 100
            },
            id: 200
        },
        name: 300
    }

    // Move deeply nested field (Level3 payload)
    var moved_payload = @root.sub.child.payload

    // Access adjacent fields in parent/sibling levels that were NOT moved
    var tag = root.sub.child.tag
    var id = root.sub.id
    var name = root.name

    if tag != 100 || id != 200 || name != 300 {
        io.PrintLn("Error in deep field partial move access")
        os.Exit(1)
    }

    io.PrintLn("Deep field partial move test passed: tag=", tag, "id=", id, "name=", name)
    return 0
}
