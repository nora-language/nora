package main

import "io"

// Simulating Option/Result Unwrap scenario
type MyOption[T] = enum {
    MyNone,
    MySome(val: T)
}

// In standard library, self is of type MyOption[T], which is owned,
// but the compiler's signature resolution registers `self` parameter with LeaseRead.
// When matched, it should NOT wrap `val` in a leased pointer type because MyOption[T] is owned.
pub fn (self: MyOption[T]) Unwrap[T]() T {
    match self {
        MySome(val) => {
            return val
        }
        MyNone => {
            panic("called Unwrap on a None value")
        }
    }
}

fn main() i32 {
    var o = MySome[i32](42)
    var val = o.Unwrap()
    io.PrintLn("Unwrapped: ${val}")
    return 0
}
