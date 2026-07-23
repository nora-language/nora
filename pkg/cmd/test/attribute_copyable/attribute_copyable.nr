package main

import "io"

[copyable]
pub type CustomPrimitive = struct {
    val: i64
}

// A generic function that requires T to be copyable
pub fn Duplicate[T: Copy](item: T) T {
    // We can use it twice because it's copyable!
    var copy1 = item
    var copy2 = copy1
    return copy2
}

pub fn main() {
    var primitive = CustomPrimitive { val: 42 }
    
    // Pass the struct into a generic function requiring [T: Copy]
    var duplicated = Duplicate[CustomPrimitive](primitive)
    
    if (duplicated.val == 42) {
        io.PrintLn("Successfully passed [copyable] struct into [T: Copy] generic!")
    }
}
