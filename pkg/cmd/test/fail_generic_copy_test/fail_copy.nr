package main

pub fn Support[T: Copy](a: T) T {
    return a
}

fn main() {
    // String is an owned type in Nora
    var s = "hello"
    // Should fail with: type 'str' is an owned type and cannot satisfy the 'Copy' constraint
    var res = Support[str](s)
}
