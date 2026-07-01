package main

import "io"

pub type Inner = struct {
    data: i32
}

pub type Dummy = struct {
    val: @Inner
}

pub fn CreateDummy() @Dummy {
    return alloc Dummy { val: alloc Inner { data: 42 } }
}

pub fn (d: #Dummy) Read() i32 {
    return d.val.data
}

fn main() i32 {
    // The allocated Dummy is an anonymous temporary.
    // The solver must track and drop the temporary after .Read() finishes.
    var val = CreateDummy().Read()
    io.PrintLn("Read value completed")
    return 0
}
