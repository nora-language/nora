package main

pub type Vector = struct {
    p0: i32,
}

pub fn NewVector(p0: i32) @Vector {
    return alloc Vector {
        p0: @p0,
    }
}

pub fn main() {
    var v = NewVector(10)
}
