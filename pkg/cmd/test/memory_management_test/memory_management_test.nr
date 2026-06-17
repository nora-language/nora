package main

import "io"

type User = struct {
    name: str,
    id: i32
}

fn consume(u: @User) {
    io.PrintLn("Consumed user ${u.name}")
}

fn display(u: #User) {
    io.PrintLn("Displaying user ${u.name} (Read-only borrow)")
}

fn main() i32 {
    // 1. Creation
    var u1 = User {
        name: "Dwi",
        id: 1
    }

    // 2. Borrowing (Read-only lease)
    display(#u1)

    // 3. Alias Tracking (Topological Dependency)
    // Using # makes name_ref an alias; u1 remains alive.
    var name_ref = #u1.name
    io.PrintLn("Accessing name via alias: ${name_ref}")

    // 4. Move Operation (Transfer ownership)
    // After this, u1 AND its alias name_ref are both invalid.
    consume(@u1)

    // 5. Violation Check (Commented out)
    // io.PrintLn("${name_ref}") // ERROR: use of borrow 'name_ref' whose origin 'u1' was moved
    io.PrintLn("Memory test sequence complete")

    return 0
}
