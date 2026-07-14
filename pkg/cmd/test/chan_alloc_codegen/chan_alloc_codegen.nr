package main

import "io"

type CommandBuffer = struct {
    handle: u64,
    id: i32
}

fn Finish(id: i32) @CommandBuffer {
    return alloc CommandBuffer {
        handle: 0x12345678,
        id: id
    }
}

fn worker(ch: #chan[CommandBuffer]) {
    var cmd = Finish(42)
    ch <- @cmd
}

fn main() i32 {
    var ch = make(chan[CommandBuffer], 1)

    scope {
        spawn worker(#ch)

        var received = <-ch
        io.PrintLn("Received id: ${received.id}")
        if (received.id != 42) {
            panic("chan_alloc_codegen_bug reproduction: received struct field 'id' corrupted")
        }
        if (received.handle != 0x12345678) {
            panic("chan_alloc_codegen_bug reproduction: received struct field 'handle' corrupted")
        }
    }

    io.PrintLn("PASS: chan_alloc_codegen test passed")
    return 0
}
