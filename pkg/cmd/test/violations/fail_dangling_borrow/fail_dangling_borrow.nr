package main

fn worker(x: #i32) {}

fn main() i32 {
    var y = 42
    var py = #y

    // 1. Should fail because worker expects a read lease (#)
    scope {
        spawn worker(#y)

        // 2. Should fail because py is a lease variable and is captured inside the spawn call
        spawn worker(py)

        return 0

    }
}
