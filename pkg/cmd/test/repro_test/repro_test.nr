package main

import "collections"
import "sync"
import "io"

[shared]
type DummyWorld = struct {
    x: i32
}

fn worker_repro(w: #DummyWorld, wg: #sync.WaitGroup) {
    io.PrintLn("Worker run: " + io.FormatInt(w.x))
    wg.Done()
}

fn test_spawn_repro(w: &DummyWorld) {
    var wg = sync.NewWaitGroup()
    wg.Add(1)
    scope {
        spawn worker_repro(#w, #wg)
        wg.Wait()

    }
}

fn test_generic_repro[T](col_any: &any, val_any: &any) {
    // 1. Vector[T] cast
    var vec: &collections.Vector[T] = collections.Vector[T](col_any)
    
    // 2. Generic T cast
    var val: &T = T(val_any)
    
    io.PrintLn("test_generic_repro called")
}

fn main() i32 {
    io.PrintLn("Gecs Repro Test")
    var col_any: any = collections.NewVector[i32](4)
    var val_any: any = 123
    test_generic_repro[i32](&col_any, &val_any)
    return 0
}
