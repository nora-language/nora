package main

import "io"

type User = struct {
    id: i32,
    name: str
}

fn (u: #User) drop() {
    io.PrintLn("Dropping User ${u.id}: ${u.name}")
}

fn main() i32 {
    io.PrintLn("--- Primitive Heap Array ---")
    {
        var arr = alloc i32[5]
        io.PrintLn("Array length: ${len(arr)}")

        arr[0] = 10
        arr[4] = 50
        io.PrintLn("arr[0] = ${arr[0]}, arr[4] = ${arr[4]}")
    }
    io.PrintLn("Primitive array freed.\n")

    io.PrintLn("--- Struct Heap Array ---")
    {
        var users = alloc User[3]
        io.PrintLn("Users length: ${len(users)}")

        users[0] = User {
            id: 1,
            name: "Alice"
        }
        users[1] = User {
            id: 2,
            name: "Bob"
        }
        users[2] = User {
            id: 3,
            name: "Charlie"
        }

        io.PrintLn("Users[1].name = ${users[1].name}")
        io.PrintLn("Leaving block...")
    }
    io.PrintLn("Struct array freed.")

    return 0
}
