package main

import "io"

type User = struct {
    name: str
}

// 1. Read borrow of an owner (#@User)
fn read_owner_borrow(u: #@User) {
    io.PrintLn("Read owner borrow: ${u.name}")
}

// 2. Mutable borrow of an owner (&@User)
fn update_owner_borrow(u: &@User) {
    u.name = "Alice Updated"
    io.PrintLn("Update owner borrow: ${u.name}")
}

// 3. Multi-level borrow (&#@User)
fn nested_owner_borrow(u: &#@User) {
    io.PrintLn("Nested owner borrow: ${u.name}")
}

fn main() i32 {
    var u1 = alloc User {
        name: "Alice"
    }
    // Test #@User
    read_owner_borrow(u1)

    var u2 = alloc User {
        name: "Alice"
    }
    // Test &@User
    update_owner_borrow(&u2)

    var u3 = alloc User {
        name: "Alice"
    }
    // Test &#@User
    nested_owner_borrow(&#u3)

    return 0
}
