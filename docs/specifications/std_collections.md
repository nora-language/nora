# Standard Collections

## Overview

The `std/collections` package provides the fundamental, heap-allocated data structures necessary for dynamic programming in Nora. Every collection fully integrates with Nora's Topological Lease Solver, ensuring that memory leaks and Use-After-Free bugs are statically prevented.

## `Vector[T]`

`Vector[T]` (`std/collections/vector.nr`) is a contiguous, heap-allocated, resizable array. It is the default collection for most sequential data.

### Instantiation
Because vectors manage heap memory, they must be created through an allocator or constructor to return an owned (`@Vector[T]`) instance:
```nora
var vec = collections.NewVector[i32](10) // 10 is initial capacity
```

### Methods
*   `Push(val: @T)`: Appends an element to the end of the vector. If the element is a heap-allocated resource (like a struct containing pointers), ownership is transferred (`@T`) into the vector. The vector takes responsibility for dropping the element when the vector is destroyed.
*   `Get(index: i32) T`: Returns a copy of the element at `index`. Ideal for primitive value-types (like `i32`).
*   `GetMut(index: i32) &T`: Returns a **mutable lease** pointing directly into the vector's memory buffer. The Lease Solver statically tracks this returned pointer, preventing the vector from being dropped, mutated, or moved while the `&T` reference is alive.
*   `GetRef(index: i32) #T`: Returns a **read-only lease** pointing to the element.
*   `Pop() Option[T]`: Removes and returns the last element, wrapped in a `Some`. Returns `None` if the vector is empty.
*   `Len() i32`: Returns the current number of elements.
*   `Clear()`: Resets the length to 0, calling the `drop()` method on all contained elements.

## `HashMap[K, V]`

`HashMap[K, V]` (`std/collections/map.nr`) provides a generic hash table implementation mapping keys to values.

### Keys
To be used as a key `K`, the type must implement the `Equatable[K]` and `Comparable[K]` interfaces from the `std/cmp` module.

### Methods
*   `Insert(key: K, val: @V)`: Inserts a key-value pair into the map. Similar to `Vector.Push`, the map takes ownership of heap-allocated values.
*   `Get(key: #K) Option[V]`: Retrieves a copy of the value associated with the key. Note that the key parameter is passed as a read-only lease (`#K`) to prevent unnecessary allocations during lookups.
*   `GetMut(key: #K) Option[&V]`: Retrieves an optional mutable lease pointing directly to the value in the map bucket.
*   `Remove(key: #K) bool`: Removes the key-value pair, calling `drop()` on the value. Returns `true` if the key existed.
*   `Contains(key: #K) bool`: Returns `true` if the key exists in the map.

## `List[T]`

`List[T]` (`std/collections/list.nr`) is a standard doubly-linked list, ideal for frequent insertions and deletions at arbitrary positions without triggering massive memory reallocations.

### Methods
*   `PushBack(val: @T)`: Appends an element to the tail.
*   `PushFront(val: @T)`: Prepends an element to the head.
*   `PopBack() Option[T]`: Removes and returns the tail.
*   `PopFront() Option[T]`: Removes and returns the head.
