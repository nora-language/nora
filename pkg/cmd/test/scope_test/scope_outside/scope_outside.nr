package main

import "sync"

fn worker(wg: #sync.WaitGroup) {
    wg.Done()
}

fn main() i32 {
    var wg = sync.NewWaitGroup()
    
    // Violation: passing exempt lease outside of scope { ... } block
    wg.Add(1)
    scope {
        spawn worker(#wg)

        wg.Wait()
        return 0

    }
}
