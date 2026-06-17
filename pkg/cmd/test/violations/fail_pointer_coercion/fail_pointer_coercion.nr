package main

type User = struct {
    id: i32
}

type Account = struct {
    balance: i32
}

fn main() i32 {
    var u = User {
        id: 42
    }
    var pu = #u
    var pa: #Account = pu // <-- SHOULD FAIL SEMANTIC ANALYZER (cannot assign #User to #Account)
    return 0
}
