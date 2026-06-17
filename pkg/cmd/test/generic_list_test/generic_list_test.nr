package main

import "io"

type Node[T] = struct {
    value: T,
    next: @Node[T]
}

fn (_head: @Node[T]) drop[T]() {
    // No manual recursive drop needed, compiler handles it!
}

fn create_node[T](val: T) @Node[T] {
    return alloc Node[T] {
        value: val,
        next: none
    }
}

// ----------------- INSERTION -----------------
fn insert_front[T](head: @Node[T], val: T) @Node[T] {
    var new_node = create_node(val)
    new_node.next = head
    return new_node
}

fn insert_back[T](head: @Node[T], val: T) @Node[T] {
    var new_node = create_node(val)
    if (head == none) {
        return new_node
    }
    var curr: &Node[T] = head
    while (curr.next != none) {
        curr = curr.next
    }
    curr.next = new_node
    return head
}

fn insert_at[T](head: @Node[T], val: T, index: i32) @Node[T] {
    if (index == 0) {
        return insert_front(head, val)
    }
    var new_node = create_node(val)
    var curr: &Node[T] = head
    var i = 0
    while (curr != none && i < index - 1) {
        curr = curr.next
        i = i + 1
    }
    if (curr != none) {
        new_node.next = curr.next
        curr.next = new_node
    }
    return head
}

fn insert_after[T](head: @Node[T], target: T, val: T) @Node[T] {
    var curr: &Node[T] = head
    while (curr != none) {
        if (curr.value == target) {
            var new_node = create_node(val)
            new_node.next = curr.next
            curr.next = new_node
            return head
        }
        curr = curr.next
    }
    return head
}

// ----------------- DELETION -----------------
fn delete_front[T](head: @Node[T]) @Node[T] {
    if (head == none) {
        return none
    }
    return head.next
}

fn delete_back[T](head: @Node[T]) @Node[T] {
    if (head == none) {
        return none
    }
    if (head.next == none) {
        return none
    }
    var curr: &Node[T] = head
    while (curr.next.next != none) {
        curr = curr.next
    }
    curr.next = none
    return head
}

fn delete_value[T](head: @Node[T], target: T) @Node[T] {
    if (head == none) {
        return none
    }
    if (head.value == target) {
        return head.next
    }
    var curr: &Node[T] = head
    while (curr.next != none) {
        if (curr.next.value == target) {
            curr.next = curr.next.next
            return head
        }
        curr = curr.next
    }
    return head
}

fn delete_at[T](head: @Node[T], index: i32) @Node[T] {
    if (head == none) {
        return none
    }
    if (index == 0) {
        return head.next
    }
    var curr: &Node[T] = head
    var i = 0
    while (curr.next != none && i < index - 1) {
        curr = curr.next
        i = i + 1
    }
    if (curr.next != none) {
        curr.next = curr.next.next
    }
    return head
}

fn delete_all[T](_head: @Node[T]) @Node[T] {
    return none
}

// ----------------- UTILS -----------------
fn print_list[T](head: #Node[T]) {
    var curr = head
    while (curr != none) {
        io.Print(curr.value)
        io.Print(" ")
        curr = curr.next
    }
    io.PrintLn("")
}

fn main() i32 {
    var list: @Node[i32] = none

    io.PrintLn("--- Insertion ---")
    list = insert_front(list, 10)
    list = insert_back(list, 20)
    list = insert_back(list, 30)
    list = insert_front(list, 5) // Expected: 5 10 20 30
    print_list(list)

    list = insert_at(list, 15, 2) // Expected: 5 10 15 20 30
    print_list(list)

    list = insert_after(list, 20, 25) // Expected: 5 10 15 20 25 30
    print_list(list)

    io.PrintLn("--- Deletion ---")
    list = delete_front(list) // Expected: 10 15 20 25 30
    print_list(list)

    list = delete_back(list) // Expected: 10 15 20 25
    print_list(list)

    list = delete_value(list, 15) // Expected: 10 20 25
    print_list(list)

    list = delete_at(list, 1) // Expected: 10 25
    print_list(list)

    list = delete_all(list) // Expected: (empty)
    io.Print("Empty: ")
    print_list(list)

    return 0
}
