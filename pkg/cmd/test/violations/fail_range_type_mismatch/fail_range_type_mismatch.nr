package main

fn main() i32 {
    // Type mismatch between range bounds (i32 vs string)
    for i in 0.."10" {
        // ...
    }
    return 0
}
