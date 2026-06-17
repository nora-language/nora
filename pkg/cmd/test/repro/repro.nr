package main

import "io"

fn test(list: &i32[]) {
    list[0] = 100
}

fn main() i32 {
    var list = [1, 2, 3]
    test(list)
    io.PrintLn(list[0])
    return 0
}
