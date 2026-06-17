# Standard Library: Random

## Overview

The `std/random` package implements pseudo-random number generators (PRNGs). By default, it provides a fast, non-cryptographically secure generator (e.g., PCG or Xoshiro256**) suitable for simulations, games, and randomized algorithms.

> [!WARNING]
> Do not use `std/random` for cryptographic purposes (such as generating session keys or tokens). Use `std/crypto/rand` instead.

## Core Types

### `Rand`
A source of random numbers. 
```nora
pub type Rand = struct {
    seed: u64
}
```

## Functions

*   `New(seed: u64) Rand`: Creates a new PRNG instance initialized with the given seed.
*   `Seed(seed: u64)`: Initializes the global shared PRNG.
*   `Int() i32`: Returns a non-negative pseudo-random 32-bit integer.
*   `Intn(n: i32) i32`: Returns a non-negative pseudo-random number in the half-open interval `[0,n)`.
*   `Float64() f64`: Returns a pseudo-random number in the half-open interval `[0.0,1.0)`.

## Thread Safety

The global generator (accessed via package-level functions like `random.Intn()`) uses a scalable spin-lock/fiber-yield mechanism to ensure safe concurrent access across fibers. For highly concurrent workloads demanding maximum performance without lock contention, instantiate a localized `Rand` struct per fiber.
