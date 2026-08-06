package main
import "io"
fn main() i32 {
    var x = 10
    var a = &x
    var b = &x
    io.PrintLn(a) // Now `a` is used here! It must live across the birth of `b`!
    return 0
}
