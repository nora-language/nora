package lsp

import (
	"fmt"
	"testing"

	"github.com/nora-language/nora/pkg/lexer"
	"github.com/nora-language/nora/pkg/parser"
	"github.com/nora-language/nora/pkg/parser/ast"
)

func TestFindNode(t *testing.T) {
	text := `package main

import "io"
import "gecs"
import "serialize"

[serialize]
type Position = struct {
    x: i32,
    y: i32
}

[serialize]
type Velocity = struct {
    dx: i32,
    dy: i32
}

type HitEvent = struct {
    damage: i32
}

fn CloneEntity(e: #gecs.Entity) gecs.Entity {
    var id_val = e.id
    var res = gecs.Entity { id: id_val }
    return res
}

fn main() {
}
`
	l := lexer.New(text, "basic.nr")
	p := parser.New(l)
	file := p.Parse("basic.nr")

	prog := &ast.Program{
		Files: []*ast.File{file},
	}

	h := NewHandler()

	// Test Line 7 (0-indexed line 6)
	fmt.Println("=== Line 7 (first serialize) ===")
	node1 := h.findNodeAt(prog, "basic.nr", Position{Line: 6, Character: 2})
	if node1 != nil {
		fmt.Printf("Node Type: %T, Pos: %v\n", node1, node1.Pos())
	} else {
		fmt.Println("No node found")
	}

	// Test Line 13 (0-indexed line 12)
	fmt.Println("=== Line 13 (second serialize) ===")
	node2 := h.findNodeAt(prog, "basic.nr", Position{Line: 12, Character: 2})
	if node2 != nil {
		fmt.Printf("Node Type: %T, Pos: %v\n", node2, node2.Pos())
	} else {
		fmt.Println("No node found")
	}
}
