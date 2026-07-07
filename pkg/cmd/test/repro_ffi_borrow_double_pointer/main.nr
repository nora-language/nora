package main

import "ffi"
import "io"

fn main() {
    var arr = alloc i32[1]
    arr[0] = 42

    // The true native C-pointer to the array elements
    var native_ptr = ptr(arr)

    // The bugged intrinsic which currently generates a double pointer (&&arr)
    var borrowed_ptr = ffi.BorrowToRaw[i32[]](#arr)
    var mut_borrowed_ptr = ffi.MutBorrowToRaw[i32[]](&arr)

    if native_ptr != borrowed_ptr {
        io.PrintLn("FAIL: ffi.BorrowToRaw generated a double-pointer instead of the native array pointer!")
        panic("Double pointer detected on BorrowToRaw!")
    }
    
    if native_ptr != mut_borrowed_ptr {
        io.PrintLn("FAIL: ffi.MutBorrowToRaw generated a double-pointer instead of the native array pointer!")
        panic("Double pointer detected on MutBorrowToRaw!")
    }

    io.PrintLn("PASS: Borrow intrinsics generated the correct single pointers!")
}