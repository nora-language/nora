package main

pub fn Support[T: Copy](a: T) T {
    // Should not consume 'a' if T is copyable
    var b = a
    var c = a // If T wasn't Copy, this would fail with 'use of moved value'
    // Use b to avoid unused variable error
    var d = c
    if false { return b }
    return d
}

pub fn SupportArithmetic[T: Copy](a: T, b: T) T {
    var offset = a - b
    var offset2 = a + b // If T wasn't Copy, this would fail because 'a' and 'b' were consumed
    // Use offset to avoid unused variable error
    if false { return offset }
    return offset2
}

fn main() {
    var x = 10.5 // f64
    var y = 20.0
    
    // Test passing primitives to [T: Copy]
    var res = Support[f64](x)
    var res2 = SupportArithmetic[f64](x, y)
    
    if false { 
        var noop = res 
        noop = res2
    }
}
