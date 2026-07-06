package main

import "std/ffi"

pub type ColorTargetState = struct {
    format: i32
}

pub type FragmentState = struct {
    targetCount: i32
    targets: ptr
}

pub fn main() {
    var color_target = alloc ColorTargetState {
        format: 1
    }

    var fragment = alloc FragmentState {
        targetCount: 1,
        targets: ffi.BorrowToRaw[ColorTargetState](#color_target)
    }
}
