package main

import "io"
import "os"

fn test_basic_range() {
    var sum = 0
    for i in 0..10 {
        sum += i
    }
    // 0+1+2+3+4+5+6+7+8+9 = 45
    if sum != 45 {
        io.PrintLn("Error: Basic range sum is wrong")
        os.Exit(1)
    }
}

fn test_negative_range() {
    var sum = 0
    for i in -5..5 {
        sum += i
    }
    // -5 + -4 + -3 + -2 + -1 + 0 + 1 + 2 + 3 + 4 = -5
    if sum != -5 {
        io.PrintLn("Error: Negative range sum is wrong")
        os.Exit(1)
    }
}

fn test_var_range() {
    var start = 10
    var end = 20
    var sum = 0
    for i in start..end {
        sum += i
    }
    // 10 + 11 + ... + 19 = 145
    if sum != 145 {
        io.PrintLn("Error: Var range sum is wrong")
        os.Exit(1)
    }
}

fn main() i32 {
    io.PrintLn("Testing Range Loops")
    test_basic_range()
    test_negative_range()
    test_var_range()
    io.PrintLn("All range loops passed!")
    return 0
}
