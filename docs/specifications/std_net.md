# Standard Library: Net

## Overview

The `std/net` package provides a portable interface for network I/O, including TCP/IP, UDP, domain name resolution, and Unix domain sockets. It is built natively on top of the Nora Fiber scheduler, ensuring that socket read/write operations yield cooperatively rather than blocking the OS thread.

## Core Interfaces

### `Conn`
```nora
pub type Conn = interface {
    io.Reader
    io.Writer
    io.Closer
    LocalAddr() Addr
    RemoteAddr() Addr
    SetDeadline(t: time.Time) Option[Error]
}
```

### `Listener`
```nora
pub type Listener = interface {
    Accept() (Conn, Option[Error])
    Close() Option[Error]
    Addr() Addr
}
```

## Functions

### `Dial(network: str, address: str) (Conn, Option[Error])`
Connects to the address on the named network.
*   Supported networks: `"tcp"`, `"tcp4"`, `"tcp6"`, `"udp"`, `"unix"`.

### `Listen(network: str, address: str) (Listener, Option[Error])`
Announces on the local network address.

## Asynchronous Fiber Integration

When `Conn.Read` or `Conn.Write` blocks (e.g., waiting for network packets), Nora's cooperative fiber runtime automatically detects `EWOULDBLOCK` / `EAGAIN` (via non-blocking socket configuration under the hood) and yields the fiber via `NR_COOPERATIVE_YIELD()`. The underlying event loop (e.g., `epoll` or `kqueue`) will resume the fiber once data is ready, resulting in highly scalable multiplexing without manual asynchronous callbacks.
