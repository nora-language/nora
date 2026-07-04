package main

pub type MyStruct = struct {
    a: i32,
    b: i32,
    ptr: ptr
}

pub extern fn TestByValue(s: MyStruct)

fn main() {
    var s = MyStruct {
        a: 42,
        b: 100,
        ptr: none
    }
    
    TestByValue(s)
}
