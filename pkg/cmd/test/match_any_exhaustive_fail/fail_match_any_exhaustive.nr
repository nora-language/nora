package main

type User = struct {
    name: str
}

fn process_any(val: any) {
    match val {
        User(u) => {
            // Missing wildcard `_ =>`
        }
    }
}

fn main() i32 {
    return 0
}
