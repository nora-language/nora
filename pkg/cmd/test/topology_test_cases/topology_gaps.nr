package test_gaps

import "io"

type Resource = struct {
    name: str
}

fn (self: #Resource) drop() {
    io.PrintLn("Dropping: ${self.name}")
}

fn test_assignment() {
    var a = Resource {
        name: "A"
    }
    var b = Resource {
        name: "B"
    }

    var x = #a
    io.PrintLn(x.name)

    x = #b
    io.PrintLn(x.name)
}

fn test_jump_break() {
    var i = 0
    while (i < 5) {
        var r = Resource {
            name: "loop_res"
        }
        if (i == 2) {
            break;
        }
        i = i + 1
    }
}

fn main() {
    io.PrintLn("--- Test Assignment ---")
    test_assignment()

    io.PrintLn("--- Test Jump Break ---")
    test_jump_break()

    io.PrintLn("Done")
}
