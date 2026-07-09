package main

import "pkg/cmd/test/repro_nested_struct_auto_drop_leak/pkg_a"
import "pkg/cmd/test/repro_nested_struct_auto_drop_leak/pkg_b"

fn main() {
    var raw = alloc pkg_a.Inner { handle: none }
    var wrap = alloc pkg_b.Inner { value: @raw }
}
