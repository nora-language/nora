package main

import "io"

type CustomList = struct {
    data: List[i32]
}

fn (c: &CustomList) Push(val: i32) {
    c.data = append(c.data, val)
}

// Implement index interface
fn (c: #CustomList) index(i: i32) i32 {
    return c.data[i]
}

// Implement index_mut interface
fn (c: &CustomList) index_mut(i: i32) &i32 {
    return &c.data[i]
}

type CustomBool = struct {
    val: bool
}

fn (c: #CustomBool) not() bool {
    if c.val {
        return false
    }
    return true
}

type CustomInt = struct {
    val: i32
}

fn (c: #CustomInt) neg() i32 {
    return 0 - c.val
}

fn (c: #CustomInt) bitnot() i32 {
    return 0 - c.val - 1
}

fn main() i32 {
    var list = alloc CustomList {
        data: make(List[i32], 0)
    }

    list.Push(10)
    list.Push(20)
    list.Push(30)

    // Test index
    if (list[0] != 10) {
        panic("list[0] != 10")
    }

    if (list[1] != 20) {
        panic("list[1] != 20")
    }

    // Test index_mut
    list[1] = 50

    if (list[1] != 50) {
        panic("list[1] != 50")
    }

    var b = alloc CustomBool { val: false }
    if (!b != true) {
        panic("!b != true")
    }

    var b2 = alloc CustomBool { val: true }
    if (!b2 != false) {
        panic("!b2 != false")
    }

    var i = alloc CustomInt { val: 42 }
    if (-i != -42) {
        panic("-i != -42")
    }

    if (~i != -43) {
        panic("~i != -43")
    }

    io.PrintLn("Success")
    return 0
}
