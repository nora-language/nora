package pkg_node

import "collections"

// Node[T] is a generic struct that contains a collections.Vector of itself.
// This is the pattern that triggers the cross-package push type mismatch bug.
pub type Node[T: Copy] = struct {
    val: T
    children_indices: @collections.Vector[i32]
}

pub type NodeTree[T: Copy] = struct {
    nodes: @collections.Vector[Node[T]]
}

pub fn NewNodeTree[T: Copy]() @NodeTree[T] {
    return alloc NodeTree[T] {
        nodes: collections.NewVector[Node[T]](16)
    }
}

// Bug: When AddNode is called from OUTSIDE pkg_node (e.g. from main),
// `node` has the cross-package monomorphized type name `pkg_node_Node_XXXXXXXX`,
// but the Vector[Node[T]] expects `Node` (local name).
// Error: "type mismatch: expected Node, got pkg_node_Node_XXXXXXXX"
pub fn (t: &NodeTree[T]) AddNode[T: Copy](node: Node[T]) {
    t.nodes.Push[Node[T]](node)
}
