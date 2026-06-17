package main

fn modify(x: #i32) {
    x = x + 1
}

fn main() i32 {
    var x = 10
    scope {
        spawn modify(#x)
        return 0

    }
}
