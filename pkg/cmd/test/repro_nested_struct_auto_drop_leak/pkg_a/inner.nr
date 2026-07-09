package pkg_a

import "io"

pub type Inner = struct {
    handle: ptr
}

pub fn (self: &Inner) drop() {
    io.PrintLn("Inner dropped")
}
