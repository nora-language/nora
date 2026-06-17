# Standard Library: Event

## Overview

The `std/event` package provides mechanisms for event-driven programming and message broadcasting. It serves as a building block for publish-subscribe patterns within Nora, often utilized in conjunction with channels and the `nursery` concurrency primitives.

## Core Types

### `Event[T]`
A generic type representing a dispatchable event payload. 

### `EventEmitter[T]`
Allows multiple listeners to subscribe to broadcasted payloads. It safely manages concurrent access to the subscriber list without heavy lock contention.

```nora
pub type EventEmitter[T] = struct {
    // internal subscriber state
}

pub fn (e: &EventEmitter[T]) Emit(payload: T)
pub fn (e: &EventEmitter[T]) Subscribe(ch: #Channel[T])
pub fn (e: &EventEmitter[T]) Unsubscribe(ch: #Channel[T])
```

## Concurrency Integration

`EventEmitter` is inherently safe for concurrent use across multiple fibers. When `Emit` is called, it performs non-blocking sends across subscriber channels, avoiding deadlocks in highly concurrent environments.
