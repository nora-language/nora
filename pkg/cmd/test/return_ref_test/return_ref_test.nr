package main

import "io"

type Counter = struct {
    value: i32
}

// Return a mutable lease &Counter
fn get_counter_ref(c: &Counter) &Counter {
    return c
}

fn get_val_ref(c: &Counter) &i32 {
    return &c.value
}

fn main() i32 {
    var c = Counter { value: 10 }
    
    // Test direct modification via returned reference
    var c_ref: &Counter = get_counter_ref(&c)
    c_ref.value = 20
    
    io.PrintLn("Value after &Counter modification: ${c.value}")
    
    // Test field reference return
    var val_ref: &i32 = get_val_ref(&c)
    val_ref = 30
    
    io.PrintLn("Value after &i32 modification: ${c.value}")
    return 0
}
