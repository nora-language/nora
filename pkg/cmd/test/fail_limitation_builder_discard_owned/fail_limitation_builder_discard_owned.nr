package main

pub type Builder = struct {
    value: i32
}

pub fn NewBuilder() @Builder {
    return alloc Builder { value: 10 }
}

pub fn main() {
    // REAL UNRESOLVED LIMITATION #2: Discarding Owned Return Values & Missing Blank Identifier (_ = expr)
    // What you want to write:
    // Calling a function/builder method that returns an owned lease @Builder without storing it: NewBuilder() or _ = NewBuilder()
    //
    // The Reality / Compiler Error:
    // Error: cannot discard owned value of type '@Builder'. it must be assigned, moved, or passed to a function to avoid memory leaks
    // Because Nora lacks a blank identifier syntax (_ = expr), discarding unassigned owned return values is forbidden.

    NewBuilder() // FAILS
}
