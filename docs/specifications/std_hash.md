# Standard Library: Hash

## Overview

The `std/hash` package provides interfaces for hash functions and implements several non-cryptographic hashing algorithms suitable for use in hash tables, Bloom filters, and other performance-critical data structures.

## Core Interface

```nora
pub type Hash = interface {
    Write(p: #[]u8) (n: i32, err: Option[Error])
    Sum(b: #[]u8) []u8
    Reset()
    Size() i32
    BlockSize() i32
}

pub type Hash32 = interface {
    Hash
    Sum32() u32
}

pub type Hash64 = interface {
    Hash
    Sum64() u64
}
```

## Built-in Algorithms

*   **`hash/fnv`**: Implements the Fowler-Noll-Vo non-cryptographic hash functions (FNV-1 and FNV-1a). Excellent for fast distribution of string keys.
*   **`hash/crc32`**: Implements the 32-bit cyclic redundancy check, or CRC-32, checksum.

## Map Integration

The Nora runtime heavily utilizes `std/hash` algorithms implicitly when `Map[K, V]` data structures are instantiated. It automatically selects the optimal hash algorithm depending on the key type `K`.
