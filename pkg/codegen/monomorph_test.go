package codegen_test

import (
	"strings"
	"testing"

	"github.com/DwiYI/Project-Nora/pkg/codegen"
	"github.com/DwiYI/Project-Nora/pkg/lexer"
	"github.com/DwiYI/Project-Nora/pkg/parser"
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/semantic"
	"github.com/DwiYI/Project-Nora/pkg/topology"
)

func TestSharedMonomorphization(t *testing.T) {
	input := `
	package main

	type Box[T] = struct {
		value: T
	}

	fn (self: #Box[T]) get[T]() T {
		return self.value
	}

	fn main() i32 {
		var b1 = Box[#str]{ value: "hello" }
		var b2 = Box[ptr]{ value: "world" }
		var x1 = b1.get()
		var x2 = b2.get()
		return 0
	}
	`

	// 1. Parse
	l := lexer.New(input, "main.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser Errors: %v", p.Errors())
	}

	// 2. Analyze
	analyzer := semantic.NewAnalyzer()
	analyzer.Loader = &MockLoader{}
	analyzer.Analyze(prog)
	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Errors: %v", analyzer.Diagnostics.ErrorMessages())
	}

	// 3. Solve Topology
	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.Solve(prog)

	// 4. Generate C Code
	gen := codegen.NewGenerator(prog, &analyzer.SemanticInfo, solver, nil, nil)
	code, err := gen.Generate()
	if err != nil {
		t.Fatalf("Codegen Failed: %v", err)
	}

	t.Logf("Generated C Code:\n%s", code)

	// 5. Assert Type-Erasure and Definition Consolidation
	// We expect exactly ONE "struct Box_ptr" definition and ONE "Box_ptr_get" method!
	if !strings.Contains(code, "struct Box_ptr {") {
		t.Errorf("Expected C output to contain struct Box_ptr definition.")
	}

	if strings.Contains(code, "struct Box_str {") || strings.Contains(code, "struct Box_ptr_ptr {") {
		t.Errorf("Unexpected un-erased C struct specialized definitions found (Box_str / Box_ptr_ptr). Type erasure failed.")
	}

	// Verify that there is exactly one shared function definition
	getDefCount := strings.Count(code, "void* Box_ptr_get(void* _env_ptr, Box_ptr* self)")
	if getDefCount != 2 { // One for prototype declaration, one for definition!
		t.Errorf("Expected exactly 2 occurrences of Box_ptr_get prototype & definition, got %d", getDefCount)
	}
}
