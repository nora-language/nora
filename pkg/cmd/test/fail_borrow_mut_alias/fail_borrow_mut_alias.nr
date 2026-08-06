package main

fn mutate_twice(a: &i32, b: &i32) {}

fn main() i32 {
    var x = 10
    // This should fail to compile: cannot borrow `x` mutably more than once at a time.
    // If the compiler compiles this, it's a critical aliasing bug.
    mutate_twice(&x, &x)
    return 0
}
