package main

import "io"
import "os"

pub type NodeData = struct {
    val: i32,
    left: @TreeNode,
    right: @TreeNode
}

pub type TreeNode = enum {
    Node(data: NodeData),
    Leaf
}

fn count_nodes(tree: #TreeNode) i32 {
    match tree {
        Node(d) => {
            return 1 + count_nodes(#d.left) + count_nodes(#d.right)
        }
        Leaf => {
            return 0
        }
    }
}

fn main() i32 {
    var l1 = alloc Leaf
    var r1 = alloc Leaf
    var root = Node(NodeData { val: 10, left: @l1, right: @r1 })

    var count = count_nodes(#root)
    if count != 1 {
        io.PrintLn("Error: Tree node count wrong")
        os.Exit(1)
    }

    io.PrintLn("Recursive sum type tree test passed: count=", count)
    return 0
}
