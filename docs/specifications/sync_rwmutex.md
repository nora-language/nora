# Sync Package: RWMutex (Specification)

## 1. Overview
The `std/sync` package provides an `RWMutex`, which is a reader/writer mutual exclusion lock. The lock can be held by an arbitrary number of readers or a single writer.

## 2. Motivation
For highly concurrent applications manipulating read-heavy shared data (like maps or caches), a standard `Mutex` causes excessive contention by blocking all operations. An `RWMutex` allows multiple read-only operations to proceed simultaneously, significantly boosting performance.

## 3. Syntax & Semantics

```nora
[shared]
pub type RWMutex = struct {
    // opaque pointer to nr_fiber_rwmutex_t
}

pub fn NewRWMutex() @RWMutex
pub fn (self: #RWMutex) RLock()
pub fn (self: #RWMutex) RUnlock()
pub fn (self: #RWMutex) Lock()
pub fn (self: #RWMutex) Unlock()
```

- **`RLock` / `RUnlock`**: Acquires/releases a shared read lock. Blocks if a writer is active.
- **`Lock` / `Unlock`**: Acquires/releases an exclusive write lock. Blocks if there are any active readers or another writer.

## 4. Example Usage

```nora
fn read_cache(rw: #sync.RWMutex) {
    rw.RLock()
    // Read from cache safely across multiple fibers
    rw.RUnlock()
}

fn write_cache(rw: #sync.RWMutex) {
    rw.Lock()
    // Mutate cache exclusively
    rw.Unlock()
}
```

## 5. Underlying Mechanism
`RWMutex` is implemented natively in the C runtime (`std/runtime/sync.c`). It uses an atomic reader count and a writer active flag. When a fiber cannot acquire the lock, it is queued into either the `readers_head` or `writers_head` linked list and explicitly yields control back to the Nora Fiber Scheduler via `park()`. This ensures no OS threads are blocked while waiting for the lock.
