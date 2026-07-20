package main

pub type Builder = struct {
    val: i32
}

pub fn NewBuilder() @Builder {
    return alloc Builder { val: 0 }
}

pub fn (b: &Builder) Add(v: i32) &Builder {
    b.val = b.val + v
    return b
}

pub fn (b: &Builder) Sub(v: i32) &Builder {
    b.val = b.val - v
    return b
}

pub fn (b: &Builder) Build() i32 {
    return b.val
}

pub fn main() {
    var b = NewBuilder()
    
    // Testing leading dot line continuation!
    var result = b
        .Add(10)
        // With a comment in between!
        .Sub(5)
        .Build()
        
    if result != 5 {
        // Just a dummy check to make sure it runs correctly
    }
}
