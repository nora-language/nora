package main
import "io"

pub fn check_len(arr: &i32[]) {
    io.PrintLn("Length is: ", len(arr))
}

pub fn main() {
    var a = alloc i32[4]
    a[0] = 1
    a[1] = 2
    check_len(&a)
}
