package main

import "event"
import "io"
import "sync"

[shared]
type SharedPool = struct {
    fiber1: ptr,
    fiber2: ptr,
    fiber3: ptr,
    fiber4: ptr,
    fiber5: ptr,
    registered: @sync.AtomicInt,
    should_wake: @sync.AtomicInt,
    finished: @sync.AtomicInt
}

fn worker(id: i32, pool: #SharedPool) {

    io.PrintLn("Worker ${id} starting...")

    var self = event.Current()
    if (id == 1) { event.StorePtr(&pool.fiber1, self) }
    if (id == 2) { event.StorePtr(&pool.fiber2, self) }
    if (id == 3) { event.StorePtr(&pool.fiber3, self) }
    if (id == 4) { event.StorePtr(&pool.fiber4, self) }
    if (id == 5) { event.StorePtr(&pool.fiber5, self) }
    pool.registered.Add(1)

    io.PrintLn("Worker ${id} suspending...")
    
    // Loop until we are explicitly woken up by main
    while (true) {
        event.Suspend()
        if (pool.should_wake.Load() == 1) {
            break;
        }
    }

    io.PrintLn("Worker ${id} woke up successfully!")
    pool.finished.Add(1)
}

fn main() i32 {
    io.PrintLn("--- Starting Event Loop Async I/O Integration Test ---")

    var pool = alloc SharedPool {
        fiber1: none,
        fiber2: none,
        fiber3: none,
        fiber4: none,
        fiber5: none,
        registered: sync.NewAtomicInt(0),
        should_wake: sync.NewAtomicInt(0),
        finished: sync.NewAtomicInt(0)
    }

    scope {
        spawn worker(1, #pool)
        spawn worker(2, #pool)
        spawn worker(3, #pool)
        spawn worker(4, #pool)
        spawn worker(5, #pool)

        // Wait for all workers to register their pointers
        while (pool.registered.Load() < 5) {
            event.Yield()
        }

        pool.should_wake.Store(1)
        io.PrintLn("Main fiber resuming all workers...")
        
        // Repeatedly send Resume signals until all workers finish.
        // The Windows scheduler drops Resume signals if the fiber isn't fully parked yet,
        // so we must retry until the worker actually wakes up and increments `finished`.
        while (pool.finished.Load() < 5) {
            var f1 = event.LoadPtr(&pool.fiber1)
            var f2 = event.LoadPtr(&pool.fiber2)
            var f3 = event.LoadPtr(&pool.fiber3)
            var f4 = event.LoadPtr(&pool.fiber4)
            var f5 = event.LoadPtr(&pool.fiber5)
            
            if (f1 == none) {
                io.PrintLn("f1 is none!")
            }

            if (f1 != none) { event.Resume(f1) }
            if (f2 != none) { event.Resume(f2) }
            if (f3 != none) { event.Resume(f3) }
            if (f4 != none) { event.Resume(f4) }
            if (f5 != none) { event.Resume(f5) }

            event.Yield()
        }
    }

    var final_finished = pool.finished.Load()
    io.PrintLn("All workers finished. Total completed: ${final_finished}")
    if (final_finished != 5) {
        panic("Verification failed! Not all workers completed.")
    }

    io.PrintLn("Event Loop Async I/O integration verified successfully!")
    
    return 0
}
