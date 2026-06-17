package main

import "io"

fn deep_call(n: i32) {
    if (n > 0) {
        deep_call(n - 1)
    } else {
        panic("depth reached")
    }
}

fn main() {
    deep_call(3)
}
