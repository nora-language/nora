package main

fn main() {
    var x = 10
    // defer must be followed by a function call, so passing an integer expression should fail
    defer x + 5
}
