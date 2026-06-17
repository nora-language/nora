package main

import "io"

type Node[T] = struct {
    value: T,
    left: @Node[T],
    right: @Node[T]
}

fn (self: @Node[i32]) insert(val: i32) @Node[i32] {
    if (self == none) {
        return alloc Node[i32] {
            value: val,
            left: none,
            right: none
        }
    }

    if (val < self.value) {
        self.left = self.left.insert(val)
    } else {
        self.right = self.right.insert(val)
    }
    return self
}

fn (self: #Node[i32]) search(val: i32) bool {
    if (self == none) {
        return false
    }
    if (self.value == val) {
        return true
    }
    if (val < self.value) {
        return self.left.search(val)
    } else {
        return self.right.search(val)
    }
}

fn (self: #Node[i32]) print_in_order() {
    if (self == none) {
        return
    }
    self.left.print_in_order()
    io.Print("${self.value} ")
    self.right.print_in_order()
}

fn (self: @Node[i32]) delete(val: i32) @Node[i32] {
    if (self == none) {
        return none
    }
    if (val < self.value) {
        self.left = self.left.delete(val)
        return @self
    } else if (val > self.value) {
        self.right = self.right.delete(val)
        return @self
    } else {
        if (self.left == none) {
            return @self.right
        } else if (self.right == none) {
            return @self.left
        }
        var succ = self.right.min_value()
        self.value = succ
        self.right = self.right.delete(succ)
        return @self
    }
}

fn (self: #Node[i32]) min_value() i32 {
    if (self.left == none) {
        return self.value
    }
    return self.left.min_value()
}

fn main() i32 {
    var bst: @Node[i32] = none
    bst = bst.insert(10)
    bst = bst.insert(5)
    bst = bst.insert(15)
    bst = bst.insert(2)
    bst = bst.insert(7)

    io.Print("Original: ")
    bst.print_in_order()
    io.PrintLn("")

    io.PrintLn("Deleting 5...")
    bst = bst.delete(5)

    io.Print("After delete: ")
    bst.print_in_order()
    io.PrintLn("")

    if (bst.search(5)) {
        io.PrintLn("Found 5 (ERROR)")
    } else {
        io.PrintLn("5 Deleted Successfully")
    }
    if (bst.search(7)) {
        io.PrintLn("7 still exists")
    }

    return 0
}
