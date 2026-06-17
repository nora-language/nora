# Standard Library: Crypto

## Overview

The `std/crypto` package implements modern cryptographic primitives, optimized for Nora's safe execution environment. The implementations are written with a focus on mitigating timing attacks (constant-time execution) and leveraging hardware acceleration where available.

## Core Modules

### `crypto/sha256`
Implements the SHA-256 hash algorithm as defined in FIPS 180-4.
```nora
import "std/crypto/sha256"

let mut hasher = sha256.New()
hasher.Write("Hello, Nora")
let sum = hasher.Sum(nil)
```

### `crypto/aes`
Implements AES encryption (Advanced Encryption Standard). Designed for direct integration with modes of operation like GCM or CBC.

### `crypto/rand`
A cryptographically secure pseudorandom number generator (CSPRNG), wrapping the host operating system's secure entropy source (e.g., `/dev/urandom` on Unix, `BCryptGenRandom` on Windows).

## Memory and Leases

Functions operating on sensitive keys or plaintext expect read-only leases (`#[]u8`) or mutable leases (`&[]u8`) to avoid unintentional copies of secret material in memory. Nora's RAII model ensures that securely zeroed memory drops predictably when sensitive buffers leave scope.
