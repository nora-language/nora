package main

import "io"
import "serialize"

[serialize]
type MyEntity = struct {
    id: i64
}

[serialize]
type MyGroup = struct {
    entities: MyEntity[]
}

fn main() i32 {
    io.PrintLn("--- Repro Serialization Cross Test ---")
    var ents = make(MyEntity[], 2)
    ents[0] = MyEntity { id: 100 }
    ents[1] = MyEntity { id: 200 }

    var g = MyGroup {
        entities: @ents
    }

    var json_str = serialize.ToJSON[MyGroup](#g)
    io.PrintLn(json_str)

    var decoded = serialize.FromJSON[MyGroup](json_str)
    io.PrintLn("Decoded entities count: ${len(decoded.entities)}")
    io.PrintLn("Decoded[0].id: ${decoded.entities[0].id}")
    io.PrintLn("Decoded[1].id: ${decoded.entities[1].id}")
    return 0
}
