package main

pub type Point = struct {
    arr: f32[3]
}

pub fn test_array_struct() {
    var p = Point{arr: [f32(1.0), f32(2.0), f32(3.0)]}
    
    // Validate initialization and read
    if p.arr[0] != f32(1.0) || p.arr[1] != f32(2.0) || p.arr[2] != f32(3.0) {
        panic("array_struct test initialization failed")
    }

    // Validate mutation
    p.arr[1] = f32(5.0)
    if p.arr[1] != f32(5.0) {
        panic("array_struct test mutation failed")
    }
}

pub type Matrix = struct {
    data: i32[4]
}

pub fn test_array_equality() {
    var a = Matrix{data: [1, 2, 3, 4]}
    var b = Matrix{data: [1, 2, 3, 4]}
    var c = Matrix{data: [1, 0, 3, 4]}
    
    if a != b {
        panic("array_equality test failed: a should equal b")
    }
    
    if a == c {
        panic("array_equality test failed: a should not equal c")
    }
}

pub type Config = struct {
    flags: i32[2]
}

pub fn check_flags(c: Config) i32 {
    return c.flags[0] + c.flags[1]
}

pub fn test_array_pass_value() {
    var c = Config{flags: [10, 20]}
    var sum = check_flags(c)
    if sum != 30 {
        panic("array_pass_value test failed")
    }
}

pub type Buffer = struct {
    data: i32[2]
}

pub fn modify_buffer(b: &Buffer) {
    b.data[0] = 100
}

pub fn test_array_pass_ptr() {
    var b = Buffer{data: [1, 2]}
    modify_buffer(&b)
    
    if b.data[0] != 100 {
        panic("array_pass_ptr test failed")
    }
}

pub type Inner = struct {
    x: i32
    y: i32
}

pub type Outer = struct {
    inners: Inner[2]
}

pub fn test_array_nested() {
    var o = Outer{inners: [Inner{x: 1, y: 2}, Inner{x: 3, y: 4}]}
    
    if o.inners[0].x != 1 || o.inners[1].y != 4 {
        panic("array_nested test failed")
    }
}

pub fn main() {
    test_array_struct()
    test_array_equality()
    test_array_pass_value()
    test_array_pass_ptr()
    test_array_nested()
}
