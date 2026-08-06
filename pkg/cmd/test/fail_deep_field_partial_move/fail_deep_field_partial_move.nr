package main

import "io"

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

    // Moving deeply nested field (Level3 payload) invalidates parent variable 'root'
    var moved_payload = @root.sub.child.payload

    // ERROR: Topological lease solver rejects accessing root or its sibling fields after partial subfield move
    var alias = root

    return 0
    return 0
}
