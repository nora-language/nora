package main
type Inner = struct { a: i32 }
type Outer = struct { inner: Inner }
fn main() {
    var o = alloc Outer { inner: Inner { a: 5 } }
}
