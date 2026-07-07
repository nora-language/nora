package main

import "ffi"
import "io"

[repr("i32")]
pub type DummyEnum = enum {
    Ok = 0
}

pub extern fn test_set_called(arg: ptr) DummyEnum
pub extern fn test_get_called() i32

fn main() {
    var ptr_arg = ffi.Malloc(4)
    
    // Call the external function. If the bug exists, this call is discarded!
    test_set_called(ptr_arg)

    // Check if the C function was actually executed
    var called = test_get_called()
    if called == 0 {
        panic("BUG: test_set_called was dropped and never executed!")
    }

    ffi.Free(ptr_arg)

    io.PrintLn("SUCCESS")
}
