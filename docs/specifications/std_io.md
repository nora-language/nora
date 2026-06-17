# Standard Library: IO

## Overview

The `std/io` package provides basic interfaces to I/O primitives. Its primary job is to wrap existing implementations of such primitives (such as those in `std/os`) into shared public interfaces that abstract away the system specifics, allowing composition through pipelines and streams.

## Core Interfaces

### `Reader`
```nora
pub type Reader = interface {
    Read(p: &[]u8) (n: i32, err: Option[Error])
}
```
Reads up to `len(p)` bytes into `p` and returns the number of bytes read.

### `Writer`
```nora
pub type Writer = interface {
    Write(p: #[]u8) (n: i32, err: Option[Error])
}
```
Writes `len(p)` bytes from `p` to the underlying data stream.

### `Closer`
```nora
pub type Closer = interface {
    Close() Option[Error]
}
```

## Combinators

Because interface values are first-class, Nora allows easily composing them:

*   `ReadCloser` (combines Reader and Closer)
*   `WriteCloser` (combines Writer and Closer)
*   `ReadWriteCloser`

## Functions

*   `Copy(dst: Writer, src: Reader) (written: i64, err: Option[Error])`: Copies from `src` to `dst` until either EOF is reached on `src` or an error occurs.
*   `ReadAll(r: Reader) ([]u8, Option[Error])`: Reads from `r` until an error or EOF and returns the data it read.

## RAII Integration
Nora's compiler ensures that variables implementing `Closer` receive automatic drop calls (`Drop()`) injected by the Topological Lease Solver at the end of their scope, implicitly calling `Close()` to prevent descriptor leaks.
