package main

pub type BigStruct = struct {
    data: i64,
    pad1: i64,
    pad2: i64,
    pad3: i64
}

pub fn main() {
    // 100,000 * 32 bytes = 3.2MB
    // This will cause a stack overflow if emitted as a C VLA, 
    // but should succeed if properly lowered to a heap allocation.
    var p_data = alloc BigStruct[100000]
    
    // Test write and read to ensure it works
    p_data[99999] = BigStruct { data: i64(42), pad1: i64(0), pad2: i64(0), pad3: i64(0) }
    
    if p_data[99999].data != i64(42) {
        // Just checking basic memory access
    }
}
