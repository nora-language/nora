# Synchronization and Communication: Channels (`chan[T]`)

## Title & Overview
While `spawn`, `scope`, and `parallel` control *how* concurrent tasks execute, **Channels (`chan[T]`)** dictate *how they communicate*. A channel is a zero-copy, typed, and thread-safe data pipe that allows one running fiber to send data securely to another fiber, eliminating the need for complex manual locking mechanisms like Mutexes.

## Motivation
Sharing memory across concurrent threads typically requires locks, which are prone to deadlocks, race conditions, and performance bottlenecks. Nora embraces the philosophy: *"Do not communicate by sharing memory; instead, share memory by communicating."* Channels provide a safe mechanism to move data ownership across the fiber boundary securely, preventing race conditions entirely at the compiler level.

## Syntax

Channels are defined by their capacity and the type of data they hold (`chan[T]`).

```nora
// 1. Creation: allocate a channel of strings with a buffer capacity of 5
var my_channel = alloc make(chan[str], 5)

// 2. Sending: Use the `<-` operator on the right side of the channel
my_channel <- "Hello, Fiber!"

// 3. Receiving: Use the `<-` operator on the left side of the channel
var message = <-my_channel
```

### 4. Multiplexing (`select`)

The `select` block allows a fiber to wait on multiple channel operations simultaneously. It blocks until one of its `case` statements can proceed, effectively multiplexing concurrency.

```nora
select {
    case v = <-c1:
        io.PrintLn("Received ${v} from c1")
    case v = <-c2:
        io.PrintLn("Received ${v} from c2")
    default:
        io.PrintLn("Neither is ready, moving on!")
}
```

## Semantics
Channels in Nora are implemented natively inside the M:N scheduler to support deep integration with fiber suspension.
1. **Send (`ch <- value`)**: When a fiber sends data into a channel, it places the data into the channel's ring buffer. If the buffer is full, the scheduler *suspends* the current fiber, freeing up the OS thread to run another fiber. The sender wakes up only when another fiber receives data, freeing space in the buffer.
2. **Receive (`<- ch`)**: When a fiber attempts to receive data, it checks the buffer. If the buffer is empty, the scheduler *suspends* the receiver fiber until another fiber sends data into the channel.
3. **Zero-Copy Transfers**: Because Nora does not use a Garbage Collector, channels do not perform deep copies of complex structs. They transfer the 8-byte pointer (lease transfer), maintaining peak C-level performance.

## Type Rules
- Channels are strongly typed. A `chan[i32]` can only send and receive 32-bit integers.
- The `make(chan[T], capacity)` function must be used with the `alloc` keyword to instantiate the channel on the heap.

## Lease Rules (The Topological Solver)
Channels exist specifically to securely cross the concurrency boundary, requiring strict ownership topology:
1. **Send Consumes Ownership**: When you send a non-primitive value into a channel (e.g., `my_channel <- @my_struct`), you must transfer ownership (`@`). The sender permanently loses access to that variable. The compiler will error if the sender attempts to use `my_struct` again.
2. **Receive Grants Ownership**: The fiber that receives the value (`var x = <-my_channel`) becomes the new, exclusive owner of the memory.
3. **Closing/Freeing**: Channels must be explicitly destroyed or garbage-collected based on ownership. If a channel is cloned across fibers, the runtime tracks its reference count.

## Examples

### Producer-Consumer Pattern
```nora
fn producer(ch: #chan[i32]) {
    for var i = 0; i < 5; i++ {
        io.PrintLn("Producing ${i}")
        ch <- i // Sends data
    }
}

fn consumer(ch: #chan[i32]) {
    for var i = 0; i < 5; i++ {
        var data = <-ch // Receives data
        io.PrintLn("Consumed ${data}")
    }
}

fn main() i32 {
    var pipe = alloc make(chan[i32], 2)
    
    // Spawn producer and consumer in the background
    scope {
        spawn producer(#pipe)
        spawn consumer(#pipe)
    }
    
    return 0
}
```

## Edge Cases
- **Deadlocks**: If a fiber attempts to receive from an empty channel and no other fiber possesses a reference to send data into it, the receiver fiber will sleep forever, causing a deadlock. (The Nora runtime detects global deadlocks if all OS threads go to sleep).
- **Zero Capacity (Unbuffered)**: If capacity is 0, the channel is "unbuffered". A sender will suspend immediately until a receiver is simultaneously ready to take the data, creating a synchronous handshake.

## Errors & Diagnostics
- **Use After Send Error**: If a developer sends ownership of data into a channel and then attempts to read or mutate the data locally, the Topological Solver will emit a `Use After Move` compiler diagnostic.
