package main

import "io"

[shared]
type SharedDatabase = struct {
    value: i32
}

type Msg = struct {
    id: i32,
    result: i32
}

fn worker(id: i32, db: #SharedDatabase, out_ch: #chan[(@Msg)]) {
    var val = db.value
    var m = alloc Msg {
        id: id,
        result: val * id
    }
    out_ch <- @m
}

fn main() i32 {
    io.PrintLn("--- Starting Actor Scheduler Channel Test ---")

    var db = alloc SharedDatabase {
        value: 10
    }

    var ch = make(chan[(@Msg)], 3)

    scope {

        spawn worker(1, #db, #ch)
        spawn worker(2, #db, #ch)
        spawn worker(3, #db, #ch)

        var sum = 0
        var i = 0
        while i < 3 {
            var msg = <-ch
            sum = sum + msg.result
            io.PrintLn("Received msg from worker: id=${msg.id}, result=${msg.result}")
            i = i + 1
        }

        io.PrintLn("Final sum (expected 60): ${sum}")
        if (sum != 60) { // 10*1 + 10*2 + 10*3 = 60
            panic("Verification failed: expected 60")
        }

        io.PrintLn("Actor channel integration test passed!")
        return 0


    }
}
