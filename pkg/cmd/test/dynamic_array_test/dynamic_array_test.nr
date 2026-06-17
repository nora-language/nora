package main

import "io"

type User = struct {
    id: int,
    name: str
}

fn (u: #User) drop() {
    // io.PrintLn("Dropping User: ${u.id}")
}

fn print_users(users: #(@User)[]) {
    var i = 0
    var n = len(users)
    while i < n {
        var u: #User = users[i]
        if (u != none) {
            io.PrintLn("  [${i}] User ID: ${u.id}, Name: ${u.name}")
        } else {
            io.PrintLn("  [${i}] (null)")
        }
        i = i + 1
    }
}

fn main() i32 {
    io.PrintLn("--- Dynamic Array (Heap) Test ---")

    // 1. Array of Primitives
    var n = 5
    var nums = alloc int[n]
    var i = 0
    while i < n {
        nums[i] = (i + 1) * 10
        i = i + 1
    }

    io.Print("Nums: ")
    i = 0
    while i < n {
        io.Print("${nums[i]} ")
        i = i + 1
    }
    io.PrintLn("")

    // 2. Array of Strings
    io.PrintLn("\n--- Array of Strings ---")
    var names = alloc str[3]
    names[0] = "Alice"
    names[1] = "Bob"
    names[2] = "Charlie"

    i = 0
    while i < 3 {
        io.PrintLn("Name ${i}: ${names[i]}")
        i = i + 1
    }

    // 3. Array of Struct Pointers (@User)
    io.PrintLn("\n--- Array of Struct Pointers ---")
    var users = alloc (@User)[3]
    users[0] = alloc User {
        id: 1,
        name: "Dave"
    }
    users[1] = alloc User {
        id: 2,
        name: "Eve"
    }
    users[2] = none

    print_users(users)

    // 4. Multi-dimensional Dynamic Array
    io.PrintLn("\n--- Multi-dimensional Array ---")
    var rows = 2
    var cols = 3
    var matrix = alloc (List[int])[rows]

    var r = 0
    while r < rows {
        matrix[r] = make(List[int], cols)
        var c = 0
        while c < cols {
            matrix[r][c] = (r * 10) + c
            c = c + 1
        }
        r = r + 1
    }

    r = 0
    while r < rows {
        io.Print("Row ${r}: ")
        var c = 0
        while c < cols {
            io.Print("${matrix[r][c]} ")
            c = c + 1
        }
        io.PrintLn("")
        r = r + 1
    }

    io.PrintLn("\nAll dynamic array tests completed successfully.")
    return 0
}
