package main

import "io"
import "sync"

type Item = struct {
    value: i32
}

fn update_item(item: &Item) {
    item.value = item.value + 1
}

fn worker(item: @Item, wg: #sync.WaitGroup) {
    update_item(&item)
    io.PrintLn("Value: " + io.FormatInt(item.value))
    wg.Done()
}

fn main() i32 {
    io.PrintLn("--- Repro Spawn Pointer Mismatch Test ---")
    var it = Item { value: 123 }
    var wg = sync.NewWaitGroup()
    wg.Add(1)
    scope {
        spawn worker(@it, #wg)
        wg.Wait()
        io.PrintLn("Success!")
        return 0

    }
}
