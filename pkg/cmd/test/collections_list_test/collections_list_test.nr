package main

import "collections"
import "io"

fn main() i32 {
    var list = collections.NewLinkedList[i32]()

    list.PushBack(20)
    list.PushFront(10)
    list.PushBack(30)

    // Expected: 10, 20, 30
    io.Print("Size: ")
    io.PrintLn(list.Len())
    io.Print("IsEmpty: ")
    if (list.IsEmpty()) {
        io.PrintLn("true")
    } else {
        io.PrintLn("false")
    }

    var curr = #list.head
    while (curr != none) {
        io.Print("Node value: ")
        io.PrintLn(curr.value)
        curr = curr.next
    }

    match list.PopFront() {
        Some(val) => {
            io.Print("Popped Front: ")
            io.PrintLn(val)
        }
        None => {
            io.PrintLn("Failed to pop front")
        }
    }

    match list.PopBack() {
        Some(val) => {
            io.Print("Popped Back: ")
            io.PrintLn(val)
        }
        None => {
            io.PrintLn("Failed to pop back")
        }
    }

    list.Clear()
    io.Print("After clear, IsEmpty: ")
    if (list.IsEmpty()) {
        io.PrintLn("true")
    } else {
        io.PrintLn("false")
    }
    io.Print("Size: ")
    io.PrintLn(list.Len())

    return 0
}
