package main

import "std/io" // <-- EXPECTED SEMANTIC ERROR: import name conflict: package "std/io" conflicts with already imported package "std/math"
import io "std/math"

fn main() i32 {
    return 0
}
