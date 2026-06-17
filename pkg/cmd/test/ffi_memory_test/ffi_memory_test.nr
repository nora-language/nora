package main

import "io"

extern fn malloc(size: i32) ptr

extern fn free(p: ptr)

type Resource = struct {
    data: ptr
}

fn (r: #Resource) drop() {
    io.PrintLn("RAII: Dropping resource and freeing memory at %p${r.data}")
    free(r.data)
}

fn test_raii() {
    io.PrintLn("Entering Scope")
    {
        var r = Resource {
            data: malloc(64)
        }
        io.PrintLn("Resource created at %p${r.data}")
        // r.drop() should be called here automatically
    }
    io.PrintLn("Exited Scope")
}

fn main() i32 {
    test_raii()
    return 0
}
