package main

import "io"

fn modify_str(s: &str) {
    s = "modified string!"
}

fn main() i32 {
    var str_val = "original string"
    io.PrintLn("Before:", str_val)
    
    modify_str(&str_val)
    io.PrintLn("After:", str_val)

    if (str_val != "modified string!") {
        return 1
    }
    return 0
}
