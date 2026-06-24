package main

import "io"

type Node[T] = struct {
    value: T,
    next: @Node[T]
}

fn create_node[T](val: T) @Node[T] {
    return alloc Node[T] {
        value: val,
        next: none
    }
}

// Method on generic struct
fn (head: @Node[T]) insert_front[T](val: T) @Node[T] {
    var new_node = create_node(val)
    new_node.next = head
    return new_node
}

fn (head: #Node[T]) print_list[T]() {
    var curr = head
    while (curr != none) {
        io.Print(curr.value)
        io.Print(" ")
        curr = curr.next
    }
    io.PrintLn("")
}



fn main() i32 {
    io.PrintLn("--- Generic Method Test ---")
    var list: @Node[i32] = create_node(10)
    list = list.insert_front(5)
    list = list.insert_front(1)

    io.Print("List: ")
    list.print_list()


    return 0
}
