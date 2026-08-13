package main

import "io"
import "collections"
import "./pkg_node"

// Repro: Cross-package generic-struct-in-collection push type mismatch.
//
// Expected: tree.AddNode[i32](n) compiles successfully.
// Actual:   Semantic error:
//   "type mismatch: expected Node, got pkg_node_Node_XXXXXXXX"
//
// Root cause: When NodeTree is monomorphized from main, the compiler names the
// Vector element type `pkg_node_Node_XXXXXXXX` (cross-package hash). But inside
// AddNode's body, nodes.Push[Node[T]](node) sees `node` as local `Node` and
// the vector expects `pkg_node_Node_XXXXXXXX`. These names don't unify.
//
// This was observed as: "type mismatch: expected BVHNode, got broad_phase_BVHNode_27d5bb29"
// in nora_physics/src/collision/broad_phase.nr

pub fn main() {
    var tree = pkg_node.NewNodeTree[i32]()
    var child_indices = collections.NewVector[i32](4)
    var n = pkg_node.Node[i32] {
        val: 42,
        children_indices: child_indices
    }
    // Bug triggers here when pkg_node.AddNode is monomorphized:
    tree.AddNode[i32](n)
    io.PrintLn("node added\n")
}
