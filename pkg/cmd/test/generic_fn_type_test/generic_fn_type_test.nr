package main

import "io"

fn test[T](val: T, h: fn(#T) i32) i32 {
    return h(val)
}

fn my_handler(val: #i32) i32 {
    return val * 2
}

fn main() i32 {
    var x = 21
    var res = test[i32](x, my_handler)
    io.PrintLn("Result: ")
    io.PrintLn(res)
    return 0
}
