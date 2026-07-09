package pkg_b

import "pkg/cmd/test/repro_nested_struct_auto_drop_leak/pkg_a"

pub type Inner = struct {
    value: @pkg_a.Inner
}
