package main

import "io"

type User = struct {
    name: str,
    age: i32
}

fn (self: #User) drop() {
    io.PrintLn("Dropping user: " + self.name)
}

fn consume_user(u: @User) {
    io.PrintLn("Consumed user: " + u.name)
    // u should be dropped here at end of function
}

fn test_loops() {
    var u = alloc User {
        name: "loop_user",
        age: 10
    }
    var i = 0
    while (i < 3) {
        io.PrintLn(u.name)
        i = i + 1
    }
    // u should be nr_free'd and User_drop'd here
}

fn test_early_return(cond: bool) @User {
    var u1 = alloc User {
        name: "first",
        age: 1
    }
    var u2 = alloc User {
        name: "second",
        age: 2
    }

    if (cond) {
        return u1 // u2 should be dropped here
    }

    return u2 // u1 should be dropped here
}

fn test_partial_move() {
    var u = User {
        name: "moved_user",
        age: 25
    }
    consume_user(@u)
    // u should NOT be dropped here because it was moved
}

fn main() i32 {
    test_loops()
    var res = test_early_return(true)
    test_partial_move()
    io.PrintLn("Done")
    return 0
}
