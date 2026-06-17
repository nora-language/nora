# Time Package: Timer and After (Specification)

## 1. Overview
The `std/time` package provides `time.Timer` and `time.After` to allow fibers to ergonomically delay execution or implement timeouts inside `select` blocks.

## 2. Motivation
While `Ticker` existed for recurring intervals, single-shot timeouts required developers to manually spawn a sleeping fiber. Providing a native `Timer` mechanism is critical for implementing resilient concurrent network and IO operations using `select` statement timeouts.

## 3. Syntax & Semantics

### `time.Timer`
A `Timer` represents a single event. When the timer expires, the current time will be sent on `Timer.C`.

```nora
pub type Timer = struct {
    C: chan[Time],
    stop_ch: chan[bool]
}

pub fn NewTimer(d: Duration) @Timer
pub fn (self: &Timer) Stop()
```

### `time.After`
A convenience function that waits for the duration to elapse and then sends the current time on the returned channel. It is equivalent to `NewTimer(d).C`.

```nora
pub fn After(d: Duration) chan[Time]
```

## 4. Example Usage

```nora
var timeout = time.After(time.Duration { nanoseconds: 50 * 1000000 })

select {
    case res = <-work_channel:
        // Handle result
    case <-timeout:
        // Timeout!
}
```

## 5. Underlying Mechanism
`Timer` spawns a detached fiber that sleeps for the specified duration and then performs a non-blocking send to the `Timer.C` channel. If `Stop()` is called, a signal is sent to `stop_ch` which terminates the sleeping fiber upon waking, ensuring the `C` channel receives nothing.
