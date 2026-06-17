package sync

import "time"

// Extern C runtime mutex bindings
extern fn nr_sync_mutex_create() ptr

extern fn nr_sync_mutex_lock(mu: ptr)

extern fn nr_sync_mutex_unlock(mu: ptr)

extern fn nr_sync_mutex_destroy(mu: ptr)

extern fn nr_sync_rwmutex_create() ptr
extern fn nr_sync_rwmutex_rlock(mu: ptr)
extern fn nr_sync_rwmutex_runlock(mu: ptr)
extern fn nr_sync_rwmutex_lock(mu: ptr)
extern fn nr_sync_rwmutex_unlock(mu: ptr)
extern fn nr_sync_rwmutex_destroy(mu: ptr)

// Extern C runtime waitgroup bindings
extern fn nr_sync_waitgroup_create() ptr

extern fn nr_sync_waitgroup_add(wg: ptr, delta: i32)

extern fn nr_sync_waitgroup_wait(wg: ptr)

extern fn nr_sync_waitgroup_destroy(wg: ptr)

// Extern C runtime atomic bindings
extern fn nr_sync_atomic_load(p: ptr) i32

extern fn nr_sync_atomic_store(p: ptr, v: i32)

extern fn nr_sync_atomic_add(p: ptr, v: i32) i32

extern fn nr_sync_atomic_cas(p: ptr, exp: ptr, des: i32) bool

// Extern safe mutators
extern fn nr_sync_add_i32(p: ptr, v: i32)

extern fn nr_sync_load_i32(p: ptr) i32

// Mutex provides mutual exclusion lock protection
[shared]
pub type Mutex = struct {
    native_handle: ptr
}

pub fn NewMutex() @Mutex {
    return alloc Mutex {
        native_handle: nr_sync_mutex_create()
    }
}

pub fn (self: #Mutex) Lock() {
    nr_sync_mutex_lock(self.native_handle)
}

pub fn (self: #Mutex) Unlock() {
    nr_sync_mutex_unlock(self.native_handle)
}

pub fn (self: #Mutex) drop() {
    nr_sync_mutex_destroy(self.native_handle)
}

// AtomicInt provides thread-safe atomic integer operations
[shared]
pub type AtomicInt = struct {
    value: i32
}

pub fn NewAtomicInt(initial: i32) @AtomicInt {
    return alloc AtomicInt {
        value: initial
    }
}

pub fn (self: #AtomicInt) Load() i32 {
    return nr_sync_atomic_load(self)
}

pub fn (self: #AtomicInt) Store(v: i32) {
    nr_sync_atomic_store(self, v)
}

pub fn (self: #AtomicInt) Add(v: i32) i32 {
    return nr_sync_atomic_add(self, v)
}

pub fn (self: #AtomicInt) CAS(expected: i32, desired: i32) bool {
    var exp = expected
    return nr_sync_atomic_cas(self, &exp, desired)
}

// AtomicBool provides thread-safe atomic boolean operations
[shared]
pub type AtomicBool = struct {
    value: i32
}

pub fn NewAtomicBool(initial: bool) @AtomicBool {
    var int_val = 0
    if (initial) {
        int_val = 1
    }
    return alloc AtomicBool {
        value: int_val
    }
}

pub fn (self: #AtomicBool) Load() bool {
    return nr_sync_atomic_load(self) != 0
}

pub fn (self: #AtomicBool) Store(v: bool) {
    var int_val = 0
    if (v) {
        int_val = 1
    }
    nr_sync_atomic_store(self, int_val)
}

pub fn (self: #AtomicBool) CAS(expected: bool, desired: bool) bool {
    var exp_int = 0
    if (expected) {
        exp_int = 1
    }
    var des_int = 0
    if (desired) {
        des_int = 1
    }
    var exp = exp_int
    return nr_sync_atomic_cas(self, &exp, des_int)
}

// Once guarantees a function block is executed exactly once
[shared]
pub type Once = struct {
    done: i32
}

pub fn NewOnce() @Once {
    return alloc Once {
        done: 0
    }
}

pub fn (self: #Once) Do(f: fn()) {
    if (nr_sync_atomic_load(self) == 1) {
        return
    }
    var expected = 0
    if (nr_sync_atomic_cas(self, &expected, 1)) {
        f()
    }
}

// WaitGroup coordinates and blocks execution until a pool of fibers completes
[shared]
pub type WaitGroup = struct {
    native_handle: ptr
}

pub fn NewWaitGroup() @WaitGroup {
    return alloc WaitGroup {
        native_handle: nr_sync_waitgroup_create()
    }
}

pub fn (self: #WaitGroup) Add(delta: i32) {
    nr_sync_waitgroup_add(self.native_handle, delta)
}

pub fn (self: #WaitGroup) Done() {
    nr_sync_waitgroup_add(self.native_handle, -1)
}

pub fn (self: #WaitGroup) Wait() {
    nr_sync_waitgroup_wait(self.native_handle)
}

pub fn (self: #WaitGroup) drop() {
    nr_sync_waitgroup_destroy(self.native_handle)
}

// RWMutex provides a reader/writer mutual exclusion lock
[shared]
pub type RWMutex = struct {
    native_handle: ptr
}

pub fn NewRWMutex() @RWMutex {
    return alloc RWMutex {
        native_handle: nr_sync_rwmutex_create()
    }
}

pub fn (self: #RWMutex) RLock() {
    nr_sync_rwmutex_rlock(self.native_handle)
}

pub fn (self: #RWMutex) RUnlock() {
    nr_sync_rwmutex_runlock(self.native_handle)
}

pub fn (self: #RWMutex) Lock() {
    nr_sync_rwmutex_lock(self.native_handle)
}

pub fn (self: #RWMutex) Unlock() {
    nr_sync_rwmutex_unlock(self.native_handle)
}

pub fn (self: #RWMutex) drop() {
    nr_sync_rwmutex_destroy(self.native_handle)
}
