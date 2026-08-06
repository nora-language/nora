package main

import "io"
import "os"

pub fn add_two(x: i32) i32 {
    return x + 2
}

pub fn mul_three(x: i32) i32 {
    return x * 3
}

pub type OperationPipeline = struct {
    op1: fn(i32) i32,
    op2: fn(i32) i32
}

fn main() i32 {
    var pipeline = OperationPipeline {
        op1: add_two,
        op2: mul_three
    }

    var step1 = pipeline.op1(5)   // 5 + 2 = 7
    var step2 = pipeline.op2(step1) // 7 * 3 = 21

    if step2 != 21 {
        io.PrintLn("Error: Function pointer field pipeline output wrong")
        os.Exit(1)
    }

    io.PrintLn("Struct function pointer field test passed: output=", step2)
    return 0
}
