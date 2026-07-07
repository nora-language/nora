package main

[repr("i32")]
pub type Color = enum { Red = 0, Green = 1 }

pub fn main() i32 {
    var c = Color.Green;
    var val: i32 = c as i32;
    var c2: Color = val as Color;
    return val;
}
