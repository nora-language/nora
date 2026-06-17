package lsp

import (
	"fmt"
	"testing"
	"github.com/DwiYI/Project-Nora/pkg/lexer"
	"github.com/DwiYI/Project-Nora/pkg/parser"
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
)

func TestDumpAst(t *testing.T) {
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

	ast.Inspect(prog, func(n ast.Node) bool {
		if n != nil && !ast.IsNil(n) {
			fmt.Printf("Type: %T, Pos: %v\n", n, n.Pos())
		}
		return true
	})
}
