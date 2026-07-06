package main

import "ffi"

fn main() {
    var p = ptr(0)
    var p2 = ffi.BorrowToRaw[ptr](#p)
}
