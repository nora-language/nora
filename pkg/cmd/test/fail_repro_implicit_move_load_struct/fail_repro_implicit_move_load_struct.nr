package main

pub type Point = struct {
    x: i32
    y: i32
}

pub fn takes_value[T](p: T) {
    // does nothing
}

pub fn main() {
    var p_val: @Point = alloc Point{x: 10, y: 20}
    takes_value[Point](p_val) // Pass WITHOUT @, this should fail semantic analysis!
}
