package main

import "io"

type User = struct {
    name: str
}

type Robot = struct {
    id: i32
}

fn process_any(val: any) {
    match val {
        User(u) => {
            io.PrintLn("Matched User")
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
    var a: any = User {
        name: "Alice"
    }
    var b: any = Robot {
        id: 101
    }

    process_any(a)
    process_any(b)

    return 0
}
