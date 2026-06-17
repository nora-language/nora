package main

import "io"

extern fn lose_it(a: #ptr)

fn main() {
    // Array allocation
    var a = [1, 2, 3]

    // Alloc expression
    var p = alloc 123

    io.PrintLn("Allocated some stuff.")

    // To leak them, I'll use a hack: move them into an extern that does nothing
    // but Nora will think they are moved and not drop them.
    lose_it(@a)
    lose_it(@p)

    io.PrintLn("Leaked via @ move to extern.")
}
