package main

import "io"

type Resource = struct {
    first: @str,
    second: @str
}

fn test_branch(r: Resource, cond: bool) @str {
    if (cond) {
        var f = @r.first
        io.PrintLn("Took first in branch:", f)
        return f
    }

    // Since the 'if' block returned early, 'r.first' is still alive when cond is false
    var f2 = @r.first
    io.PrintLn("Took first outside branch:", f2)
    return f2
}

fn main() i32 {
    var r1 = Resource { first: "alpha", second: "beta" }
    var res1 = test_branch(r1, true)
    
    var r2 = Resource { first: "gamma", second: "delta" }
    var res2 = test_branch(r2, false)

    io.PrintLn("Results:", res1, res2)
    return 0
}
