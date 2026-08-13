package main

import "io"

pub type PointerWrapper = struct {
    handle: ptr
}

pub fn (self: &PointerWrapper) drop() {
    io.PrintLn("PointerWrapper.drop() executed!")
}

pub type OuterStruct = struct {
    inner: @PointerWrapper
}

pub fn createOuter() @OuterStruct {
    var inner = alloc PointerWrapper { handle: none }
    return alloc OuterStruct { inner: inner }
}

pub fn main() i32 {
    io.PrintLn("Starting test...")
    {
        var outer = createOuter()
        // outer.inner should be dropped automatically when leaving this scope!
    }
    io.PrintLn("Test finished. If drop was not called, this is a bug!")
    return 0
}
