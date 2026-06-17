package main

import "io"

pub fn main() {
    var arr = alloc i32[10]
    
    // Initialize
    var i = 0
    while i < len(arr) {
        arr[i] = i
        i++
    }

    var sum = 0
    // Test BCE optimization: No bounds check should be generated for this standard loop!
    i = 0
    while i < len(arr) {
        sum = sum + arr[i]
        i = i + 1
    }

    // Test BCE negative case: i is mutated, should fallback to bounds check
    i = 0
    while i < len(arr) {
        if (i == 5) {
            i = 6 // Mutation! Bounds check must NOT be eliminated.
        }
        sum = sum + arr[i]
        i++
    }

    io.PrintLn("Success")
}
