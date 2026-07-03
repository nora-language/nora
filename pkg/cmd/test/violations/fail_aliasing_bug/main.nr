package main

import "pkg/cmd/test/violations/fail_aliasing_bug/my_pkg"

pub type MyDummy = my_pkg.Dummy
pub type MyInt = i32
pub type MyHandle = ptr

pub fn foo(d: MyDummy, i: MyInt) {
}

pub extern fn do_something(handle: MyHandle) MyHandle

fn main() i32 {
    return 0
}
