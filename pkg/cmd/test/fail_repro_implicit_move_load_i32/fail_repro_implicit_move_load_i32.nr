package main

pub fn takes_value_i32(p: i32) {
    // does nothing
}

pub fn main() {
    var p_val: @i32 = alloc 42
    takes_value_i32(p_val)
}
