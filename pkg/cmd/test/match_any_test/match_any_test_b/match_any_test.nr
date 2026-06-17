package main

import "io"

type User = struct {
    name: str
}

type Robot = struct {
    id: i32
}

fn test_match_any(val: any) {
    match val {
        User(u) => {
            io.PrintLn("Matched User: " + u.name)
        }
        Robot(r) => {
            io.PrintLn("Matched Robot")
        }
        _ => {
            io.PrintLn("Matched unknown")
        }
    }
}

fn main() i32 {
    var u = User{name: "Alice"}
    var a: any = #u
    test_match_any(a)
    return 0
}
