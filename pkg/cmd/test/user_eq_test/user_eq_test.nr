package main

import "io"

type User = struct {
    id: i32,
    name: str
}

// Custom equality: only compare IDs
fn (self: #User) eq(other: #User) bool {
    io.PrintLn("Using custom EQ for ${self.name} and ${other.name}")
    return self.id == other.id
}

fn main() i32 {
    var u1 = User {
        id: 1,
        name: "Alice"
    }
    var u2 = User {
        id: 1,
        name: "Bob"
    }
    var u3 = User {
        id: 2,
        name: "Alice"
    }

    io.PrintLn("Testing Custom Equality:")
    if (u1 == u2) {
        io.PrintLn("u1 == u2: PASSED (IDs match)")
    } else {
        io.PrintLn("u1 == u2: FAILED")
        return 1
    }

    if (u1 != u3) {
        io.PrintLn("u1 != u3: PASSED (IDs differ)")
    } else {
        io.PrintLn("u1 != u3: FAILED")
        return 1
    }

    io.PrintLn("\nCustom Equality test passed!")
    return 0
}
