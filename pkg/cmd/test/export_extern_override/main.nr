package main

import "./lib"

extern fn test_export_function() i32

pub fn main() {
    test_export_function()
}
