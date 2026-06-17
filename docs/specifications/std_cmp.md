# Standard Comparators

## Overview

The `std/cmp` module establishes the core foundational interfaces required for value comparison in Nora. By implementing these interfaces, custom structs and types can integrate deeply with the standard library, enabling operations like HashMap key indexing, sorting, and standard equality checks.

## `Equatable[T]`

The `Equatable` interface requires a type to implement an `eq` method.

```nora
pub type Equatable[T] = interface {
    fn eq(other: #T) bool
}
```

### Usage
Types that implement `Equatable[T]` can be compared for logical equality. This is fundamentally different from reference equality (checking if two pointers point to the exact same memory address). 

When a type implements `Equatable[T]`, the Nora compiler allows it to be used as a Key `K` in `collections.HashMap[K, V]`, as the map must be able to resolve hash collisions by verifying logical equality between keys.

## `Comparable[T]`

The `Comparable` interface requires a type to implement a `cmp` method, defining a total ordering over the type.

```nora
pub type Comparable[T] = interface {
    fn cmp(other: #T) i32
}
```

### Semantics
The `cmp` method must return:
- `< 0` (typically `-1`) if `self` is logically strictly less than `other`.
- `0` if `self` is logically equal to `other`.
- `> 0` (typically `1`) if `self` is logically strictly greater than `other`.

### Usage
Types implementing `Comparable[T]` can utilize the standard library sorting algorithms (`std/collections/sort`) and can be inserted into ordered data structures (like standard Binary Search Trees or Priority Queues).
