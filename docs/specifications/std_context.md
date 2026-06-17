# Standard Library: Context

## Overview

The `std/context` package provides the `Context` interface and related utilities for carrying deadlines, cancellation signals, and other request-scoped values across API boundaries and between fibers. It is heavily inspired by Go's `context` package, adapted for Nora's `Topological Lease Solver` and zero-cost abstraction model.

## Core Types

### The Context Interface
The core of the module is the `Context` interface, which provides methods to check for cancellation or deadlines.

```nora
pub type Context = interface {
    Deadline() (Time, bool)
    Done() #Channel[struct{}]
    Err() Option[Error]
    Value(key: any) Option[any]
}
```

## Functions

### `Background() Context`
Returns a non-nil, empty Context. It is never canceled, has no values, and has no deadline. It is typically used by the main function, initialization, and tests, and as the top-level Context for incoming requests.

### `WithCancel(parent: Context) (Context, CancelFunc)`
Returns a copy of the parent with a new Done channel. The returned context's Done channel is closed when the returned cancel function is called or when the parent context's Done channel is closed, whichever happens first.

### `WithDeadline(parent: Context, d: Time) (Context, CancelFunc)`
Returns a copy of the parent context with the deadline adjusted to be no later than `d`.

### `WithValue(parent: Context, key: any, val: any) Context`
Returns a copy of the parent in which the value associated with key is `val`.

## Memory & Lease Management

Contexts are frequently passed as read-only leases (`#Context`) to avoid unnecessary ownership transfers, ensuring the topological solver handles their lifetimes predictably without extraneous runtime overhead.

```nora
pub fn HandleRequest(ctx: #Context, req: Request) {
    // ...
}
```
