package main

import "ffi"
import "io"

fn main() {
    var arr = alloc i32[1]
    var p: ptr = ptr(arr)

    var p_addr = ffi.BorrowToRaw[ptr](#p)

    if p_addr == p {
        panic("BUG: ffi.BorrowToRaw[ptr](#p) stripped the address-of operator!")
    }

    var p_mut_addr = ffi.MutBorrowToRaw[ptr](&p)
    if p_mut_addr == p {
        panic("BUG: ffi.MutBorrowToRaw[ptr](&p) stripped the address-of operator!")
    }

    io.PrintLn("PASS") // Invalidate cache 4
}
