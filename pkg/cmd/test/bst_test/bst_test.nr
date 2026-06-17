package main

import "io"

type Node[T] = struct {
    value: T,
    left: @Node[T],
    right: @Node[T]
}

fn insert(root: @Node[i32], val: i32) @Node[i32] {
    if (root == none) {
        return alloc Node[i32] {
            value: val,
            left: none,
            right: none
        }
    }

    if (val < root.value) {
        root.left = insert(@root.left, val)
    } else {
        root.right = insert(@root.right, val)
    }
    return root
}

fn search(root: #Node[i32], val: i32) bool {
    if (root == none) {
        return false
    }
    if (root.value == val) {
        return true
    }
    if (val < root.value) {
        return search(#root.left, val)
    } else {
        return search(#root.right, val)
    }
}

fn print_in_order(root: #Node[i32]) {
    if (root == none) {
        return
    }
    print_in_order(#root.left)
    io.Print("${root.value}")
    io.Print(" ")
    print_in_order(#root.right)
}

fn delete(root: @Node[i32], val: i32) @Node[i32] {
    if (root == none) {
        return none
    }
    if (val < root.value) {
        root.left = delete(@root.left, val)
        return @root
    } else if (val > root.value) {
        root.right = delete(@root.right, val)
        return @root
    } else {
        if (root.left == none) {
            return @root.right
        } else if (root.right == none) {
            return @root.left
        }
        var succ = min_value(#root.right)
        root.value = succ
        root.right = delete(@root.right, succ)
        return @root
    }
}

fn min_value(root: #Node[i32]) i32 {
    if (root.left == none) {
        return root.value
    }
    return min_value(#root.left)
}

fn main() i32 {
    var bst = insert(none, 10)
    bst = insert(@bst, 5)
    bst = insert(@bst, 15)
    bst = insert(@bst, 2)
    bst = insert(@bst, 7)

    io.Print("Original: ")
    print_in_order(#bst)
    io.PrintLn("")

    io.PrintLn("Deleting 5...")
    bst = delete(@bst, 5)

    io.Print("After delete: ")
    print_in_order(#bst)
    io.PrintLn("")

    if (search(#bst, 5)) {
        io.PrintLn("Found 5 (ERROR)")
    } else {
        io.PrintLn("5 Deleted Successfully")
    }
    if (search(#bst, 7)) {
        io.PrintLn("7 still exists")
    }

    return 0
}
