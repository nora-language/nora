package main

pub type Vec[T: Copy] = struct {
    x: T
}

pub fn NewVec[T: Copy](x: T) Vec[T] {
    return Vec[T] { x: x }
}

pub fn (self: #Vec[T]) operator-(other: Vec[T]) Vec[T] {
    return Vec[T] { x: self.x - other.x }
}

pub fn (self: #Vec[T]) TestSub(other: Vec[T]) Vec[T] {
    // Passing self (a lease #Vec[T]) as argument to 'other' parameter (which is Vec[T] by value in Nora, but passed by pointer in C)
    return self - other
}

pub fn main() i32 {
    var a = NewVec[i32](10)
    var b = NewVec[i32](3)
    var res = a.TestSub(b)
    if (res.x != 7) {
        panic("TestSub failed")
    }
    return 0
}
