package main

import "sync"
import "time"
import "io"

fn reader_worker(rw: #sync.RWMutex, count: #sync.AtomicInt, wg: #sync.WaitGroup) {
    rw.RLock()
    count.Add(1)
    time.Sleep(time.Duration { nanoseconds: 50 * 1000000 })
    rw.RUnlock()
    wg.Done()
}

fn writer_worker1(rw: #sync.RWMutex, writer_active: #sync.AtomicBool, wg: #sync.WaitGroup) {
    rw.Lock()
    writer_active.Store(true)
    time.Sleep(time.Duration { nanoseconds: 50 * 1000000 })
    writer_active.Store(false)
    rw.Unlock()
    wg.Done()
}

fn writer_worker2(rw: #sync.RWMutex, writer_active: #sync.AtomicBool, wg: #sync.WaitGroup) {
    rw.Lock()
    if (writer_active.Load()) {
        panic("RWMutex failed: two writers active at same time!")
    }
    writer_active.Store(true)
    time.Sleep(time.Duration { nanoseconds: 50 * 1000000 })
    writer_active.Store(false)
    rw.Unlock()
    wg.Done()
}

fn main() i32 {
    io.PrintLn("=== Testing sync.RWMutex ===")
    
    var rw = sync.NewRWMutex()
    var count = sync.NewAtomicInt(0)
    var wg = sync.NewWaitGroup()
    
    wg.Add(10)
    
    // Start 10 readers
    var i = 0
    while (i < 10) {
        spawn reader_worker(#rw, #count, #wg)
        i = i + 1
    }
    
    time.Sleep(time.Duration { nanoseconds: 10 * 1000000 })
    // All 10 readers should be active simultaneously
    if (count.Load() != 10) {
        panic("RWMutex failed to allow multiple readers")
    }
    
    wg.Wait()
    io.PrintLn("Readers test passed.")
    
    // Writer test
    wg.Add(2)
    var writer_active = sync.NewAtomicBool(false)
    
    spawn writer_worker1(#rw, #writer_active, #wg)
    
    time.Sleep(time.Duration { nanoseconds: 10 * 1000000 })
    
    spawn writer_worker2(#rw, #writer_active, #wg)
    
    wg.Wait()
    io.PrintLn("Writers test passed.")
    
    io.PrintLn("All RWMutex tests passed.")
    return 0
}
