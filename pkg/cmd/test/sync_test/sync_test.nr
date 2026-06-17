package main

import "io"
import "sync"

[shared]
type SharedData = struct {
    counter: i32
}

extern fn nr_sync_add_i32(p: ptr, v: i32)

extern fn nr_sync_load_i32(p: ptr) i32

var once_counter = 0

fn increment_once() {
    once_counter = once_counter + 1
}

fn mutex_worker(data: #SharedData, wg: #sync.WaitGroup, mutex: #sync.Mutex, once: #sync.Once) {
    var i = 0
    while (i < 100) {
        mutex.Lock()
        nr_sync_add_i32(data, 1)
        mutex.Unlock()

        once.Do(increment_once)
        i = i + 1
    }
    wg.Done()
}

fn atomic_worker(atom: #sync.AtomicInt, wg: #sync.WaitGroup) {
    var i = 0
    while (i < 100) {
        atom.Add(1)
        i = i + 1
    }
    wg.Done()
}

fn main() i32 {
    io.PrintLn("--- Starting Sync Primitives Integration Test ---")

    var mutex = sync.NewMutex()
    var once = sync.NewOnce()
    var data = alloc SharedData {
        counter: 0
    }
    var wg = sync.NewWaitGroup()
    wg.Add(5)

    // Spawn 5 concurrent fibers using mutex
    scope {
        spawn mutex_worker(#data, #wg, #mutex, #once)
        spawn mutex_worker(#data, #wg, #mutex, #once)
        spawn mutex_worker(#data, #wg, #mutex, #once)
        spawn mutex_worker(#data, #wg, #mutex, #once)
        spawn mutex_worker(#data, #wg, #mutex, #once)

        wg.Wait()

        var final_counter = nr_sync_load_i32(data)
        io.PrintLn("Mutex counter final value (expected 500): ${final_counter}")
        if (final_counter != 500) {
            panic("Mutex verification failed!")
        }

        io.PrintLn("Once counter final value (expected 1): ${once_counter}")
        if (once_counter != 1) {
            panic("Once verification failed!")
        }

        // Atomic tests
        var atom = sync.NewAtomicInt(0)
        var atom_wg = sync.NewWaitGroup()
        atom_wg.Add(5)

        spawn atomic_worker(#atom, #atom_wg)
        spawn atomic_worker(#atom, #atom_wg)
        spawn atomic_worker(#atom, #atom_wg)
        spawn atomic_worker(#atom, #atom_wg)
        spawn atomic_worker(#atom, #atom_wg)

        atom_wg.Wait()

        var final_atom = atom.Load()
        io.PrintLn("AtomicInt final value (expected 500): ${final_atom}")
        if (final_atom != 500) {
            panic("AtomicInt verification failed!")
        }

        // Atomic CAS test
        var expected = 500
        var res = atom.CAS(expected, 1000)
        io.PrintLn("AtomicInt CAS success (expected true): ${res}")
        if (!res) {
            panic("Atomic CAS verification failed!")
        }

        var after_cas = atom.Load()
        io.PrintLn("AtomicInt after CAS value (expected 1000): ${after_cas}")
        if (after_cas != 1000) {
            panic("Atomic CAS value verification failed!")
        }

        // Atomic Bool test
        var b = sync.NewAtomicBool(false)
        if (b.Load()) {
            panic("AtomicBool load initial value failed!")
        }
        b.Store(true)
        if (!b.Load()) {
            panic("AtomicBool load stored value failed!")
        }

        var res_b = b.CAS(true, false)
        if (!res_b) {
            panic("AtomicBool CAS success failed!")
        }
        if (b.Load()) {
            panic("AtomicBool after CAS value failed!")
        }

        io.PrintLn("All safe synchronization primitives verified successfully!")
        return 0

    }
}

