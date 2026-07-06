package main

[NoEmit]
pub type div_t = struct {
    quot: i32,
    rem: i32
}

[NoEmit]
pub extern fn rand() i32

fn main() {
    var x = rand()
}
