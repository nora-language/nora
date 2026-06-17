package main

import "io"
import "serialize"

[serialize]
type DataObj = struct {
    values: i32[]
}

type Processor = struct {
    process1: fn(str) @DataObj,
    process2: fn(str) @DataObj
}

fn test_multiple_lambdas() {
    // Scenario 1: Multiple lambdas defined inside a single statement (struct literal)
    var p = Processor {
        process1: fn(s: str) @DataObj {
            var res = serialize.FromJSON[DataObj](s)
            return @res
        },
        process2: fn(s: str) @DataObj {
            var res = serialize.FromJSON[DataObj](s)
            return @res
        }
    }

    var json_str1 = "{\"values\":[10, 20]}"
    var json_str2 = "{\"values\":[10, 20]}"
    
    var obj1 = p.process1(json_str1)
    var obj2 = p.process2(json_str2)

    io.PrintLn("Obj1 val: ${obj1.values[0]}")
    io.PrintLn("Obj2 val: ${obj2.values[0]}")
}

fn test_mixed_recursive_blocks(cond: bool) {
    // Scenario 2: A lambda and an IfExpression inside the same statement
    var lambda_var = fn() {
        var res = serialize.FromJSON[DataObj]("{\"values\":[100]}")
        io.PrintLn("Lambda inside statement: ${res.values[0]}")
    }

    var x = if cond {
        lambda_var()
        1
    } else {
        2
    }
    
    io.PrintLn("Result x: ${x}")
}

fn main() i32 {
    io.PrintLn("Starting Multi-Lambda Scenario Tests...")
    test_multiple_lambdas()
    test_mixed_recursive_blocks(true)
    return 0
}
