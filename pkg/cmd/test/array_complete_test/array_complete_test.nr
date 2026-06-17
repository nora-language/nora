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
    io.PrintLn("--- Array Complete Implementation Test ---")

    // 1. Multi-dimensional Alloc
    io.PrintLn("\n1. Multi-dim Alloc:")
    var matrix = alloc i32[2][3]
    io.PrintLn("Outer matrix length: ${len(matrix)}")
    matrix[0] = alloc i32[2]
    matrix[1] = alloc i32[2]
    matrix[0][1] = 42
    matrix[1][0] = 100
    io.PrintLn("matrix[0][1] = ${matrix[0][1]}, matrix[1][0] = ${matrix[1][0]}")

    // 2. String literal copy on assignment (RAII Safety)
    io.PrintLn("\n2. RAII String Copy:")
    {
        var users = alloc User[2]
        users[0] = User { // Should copy "Alice"
            id: 1,
            name: "Alice"
        }
        users[1] = User { // Should copy "Bob"
            id: 2,
            name: "Bob"
        }
        io.PrintLn("User 0: ${users[0].name}, User 1: ${users[1].name}")
        io.PrintLn("Leaving inner block (should drop users safely)...")
    }
    io.PrintLn("Returned from inner block.")

    // 3. Array Slicing
    io.PrintLn("\n3. Array Slicing:")
    var arr = alloc i32[5]
    arr[0] = 10
    arr[1] = 20
    arr[2] = 30
    arr[3] = 40
    arr[4] = 50

    var sub = arr[1:4]
    io.PrintLn("Slice arr[1:4] length: ${len(sub)}")
    io.PrintLn("sub[0]=${sub[0]}, sub[1]=${sub[1]}, sub[2]=${sub[2]}")

    var head = arr[:2]
    io.PrintLn("Slice arr[:2] length: ${len(head)}")
    io.PrintLn("head[0]=${head[0]}, head[1]=${head[1]}")

    var tail = arr[3:]
    io.PrintLn("Slice arr[3:] length: ${len(tail)}")
    io.PrintLn("tail[0]=${tail[0]}, tail[1]=${tail[1]}")

    // 4. String Slicing
    io.PrintLn("\n4. String Slicing:")
    var s = "Hello Nora"
    var world = s[6:13]
    io.PrintLn("Slicing '${s}'[6:13] = '${world}'")

    return 0
}
