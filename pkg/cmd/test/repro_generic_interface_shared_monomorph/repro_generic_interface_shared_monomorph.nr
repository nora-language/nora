package main

import "io"

type Identifiable = interface {
    fn get_id() i32
}

type User = struct {
    id: i32
}

fn (u: #User) get_id() i32 {
    return u.id
}

type Device = struct {
    serial: i32
}

fn (d: #Device) get_id() i32 {
    return d.serial
}

fn fetch_id[T: Identifiable](item: #T) i32 {
    return item.get_id()
}

fn main() i32 {
    var u = User { id: 42 }
    var d = Device { serial: 101 }

    var id1 = fetch_id(#u)
    var id2 = fetch_id(#d)

    io.PrintLn("User ID:", id1)
    io.PrintLn("Device ID:", id2)

    if (id1 != 42 || id2 != 101) {
        return 1
    }
    return 0
}
