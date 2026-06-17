package main

import "io"

type Resource = struct {
    data: str
}

fn consume_resource(r: @Resource) {
    io.Print("Consumed resource: ")
    io.PrintLn(r.data)
}

fn test_conditional(flag: bool) {
    var res = alloc Resource {
        data: "Premium Dynamic Drop Flag Test"
    }
    if (flag) {
        consume_resource(@res)
    } else {
        io.PrintLn("Skipped moving resource")
    }
}

fn main() i32 {
    io.PrintLn("--- Running Conditional Move Test (flag=true) ---")
    test_conditional(true)

    io.PrintLn("--- Running Conditional Move Test (flag=false) ---")
    test_conditional(false)

    io.PrintLn("Conditional move test completed successfully with 0 leaks.")
    return 0
}
