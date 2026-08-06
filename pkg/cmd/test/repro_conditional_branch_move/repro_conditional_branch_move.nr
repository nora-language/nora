package main

import "io"

pub type HeapResource = struct {
    data: @i32[]
}

fn consume(r: @HeapResource) {
    io.PrintLn("Consumed resource in if branch")
}

fn test_conditional_move(flag: bool) {
    var arr = alloc i32[10]
    var res = HeapResource { data: @arr }

    if flag {
        // Consumed in if branch
        consume(@res)
    } else {
        // Not consumed in else branch: topological solver must drop res at block end
        io.PrintLn("Keeping resource in else branch")
    }
}

fn main() i32 {
    test_conditional_move(true)
    test_conditional_move(false)
    io.PrintLn("Conditional move test completed successfully")
    return 0
}
