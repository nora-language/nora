package main

import "io"

pub type SimpleBox = struct {
    data: @i32[]
}

fn main() i32 {
    var opt: Option[SimpleBox] = None[SimpleBox]

    for i in 0..10 {
        match opt {
            Some(b) => {
                // Extracted previous box payload
            }
            None => {}
        }
        // Re-assigning opt inside loop iteration
        var new_arr = alloc i32[10]
        opt = Some[SimpleBox](SimpleBox { data: @new_arr })
    }

    io.PrintLn("Option loop re-assignment completed safely")
    return 0
}
