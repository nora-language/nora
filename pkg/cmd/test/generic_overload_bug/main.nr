package main

pub type Box[T] = struct {
    val: T
}

// First overload: Box[T] + Box[T]
pub fn (_self: #Box[T]) operator+(other: Box[T]) Box[T] {
    return Box[T] { val: other.val } // Dummy implementation
}

// Second overload: Box[T] + T
pub fn (_self: #Box[T]) operator+(scalar: T) Box[T] {
    return Box[T] { val: scalar } // Dummy implementation
}

pub fn main() {
    var b1 = Box[i32] { val: 10 }
    var b2 = Box[i32] { val: 20 }
    
    // Test 1: Box + Box
    var b3 = b1 + b2
    if b3.val != 20 { panic("Box + Box failed") }
    
    // Test 2: Box + Scalar
    var b4 = b1 + 5
    if b4.val != 5 { panic("Box + Scalar failed") }
    // Test 3: Box + Box again to verify `st.Methods` is not corrupted
    var _b5 = b1 + b2
}
