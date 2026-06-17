package main

import "io"

type User = struct {
    id: i32,
    name: str
}

fn (u: #User) drop() {
    io.PrintLn("Dropping heap user ${u.id}: ${u.name}")
}

fn process_user(u: #User) {
    io.PrintLn("Inside process_user: ${u.id}")
    // u should be dropped here!
}

fn main() i32 {
    io.PrintLn("Start main")

    {
        io.PrintLn("Entering block")
        var u = alloc User {
            id: 7,
            name: "Heap User"
        }
        io.PrintLn("User name before move: ${u.name}")

        process_user(#u)

        // io.PrintLn("User name after move: ${u.name}") // This should be a semantic error!
        io.PrintLn("Leaving block")
    }

    io.PrintLn("End main")
    return 0
}
