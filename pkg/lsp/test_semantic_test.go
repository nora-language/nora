package lsp

import (
	"fmt"
	"testing"
	"github.com/DwiYI/Project-Nora/pkg/lexer"
	"github.com/DwiYI/Project-Nora/pkg/parser"
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/semantic"
)

func TestSemanticTokens(t *testing.T) {
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
`
	l := lexer.New(text, "basic.nr")
	p := parser.New(l)
	file := p.Parse("basic.nr")

	prog := &ast.Program{
		Files: []*ast.File{file},
	}

    analyzer := semantic.NewAnalyzer()
    analyzer.Analyze(prog)

    ast.Inspect(prog, func(n ast.Node) bool {
		if n == nil || ast.IsNil(n) {
			return false
		}
        fmt.Printf("Visited node %T at %v\n", n, n.Pos())
		if n.Pos().Filename != file.Name {
            fmt.Printf("  SKIPPED due to filename mismatch: %q vs %q\n", n.Pos().Filename, file.Name)
			return true // Continue to next node, but skip this one
		}
        if ident, ok := n.(*ast.Identifier); ok {
            fmt.Printf("  Identifier: %s\n", ident.Value)
        }
        return true
    })
}
