package main

import "io"

type Greeter = interface {
    fn greet() void
}

type User = struct {
    name: str
}

fn (u: #User) greet() void {
    io.PrintLn("Hello from User ${u.name}")
}

type Robot = struct {
    id: i32
}

fn (r: #Robot) greet() void {
    io.PrintLn("Beep Boop from Robot ${r.id}")
}

fn main() i32 {
    var list: Greeter[] = make(Greeter[], 0)

    var u = User {
        name: "Alice"
    }
    var r = Robot {
        id: 42
    }

    list = append(list, u)
    list = append(list, r)

    for g in list {
        g.greet()
    }
    return 0
}
