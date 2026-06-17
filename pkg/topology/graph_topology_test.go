package topology_test

import (
	"testing"

	"github.com/DwiYI/Project-Nora/pkg/lexer"
	"github.com/DwiYI/Project-Nora/pkg/parser"
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/semantic"
	"github.com/DwiYI/Project-Nora/pkg/topology"
)

// TestGraphTopologyDependencies: verifies that when a borrow of an owned variable
// is used, the provider's lifecycle is extended to match the borrow's last use.
//
// u is an owned struct (Node). n = #u (borrow). When n is last used at index 2,
// u's lifecycle must also extend to index 2 → u is dropped there.
// n is #Node (borrow, not owned) → no drop for n itself.
func TestGraphTopologyDependencies(t *testing.T) {
	input := `
    package main
    type Node = struct { val: i32 }
	extern fn pass_node(a: #Node)
	fn main() {
		var u = Node{ val: 100 }   // index 0: u is Node (owned)
		var n = #u                 // index 1: n borrows u → n depends on u
		pass_node(n)               // index 2: last use of n (and thus u)
		pass_node(u)               // index 3
	}
	`
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser Errors: %v", p.Errors())
	}

	analyzer := semantic.NewAnalyzer()
	analyzer.Analyze(prog)
	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Errors: %v", analyzer.Diagnostics.ErrorMessages())
	}

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.Solve(prog)

	var mainBlock *ast.BlockStatement
	for _, stmt := range file.Statements {
		if fn, ok := stmt.(*ast.FunctionStatement); ok && fn.Name.Value == "main" {
			mainBlock = fn.Body
			break
		}
	}
	if mainBlock == nil {
		t.Fatal("Could not find main function")
	}

	drops := solver.Drops[mainBlock]
	// fmt.Printf("DEBUG: Graph Drops: %+v\n", drops)

	// u is owned (Node struct). n = #u (borrow) depends on u.
	// n's last use is at index 2 → u's lifecycle extends to index 2 → u dropped after its actual last use index (3) → 4.
	checkDrop(t, drops, 4, "u", true)

	// n is a borrow (#Node), NOT owned → no drop for n
	checkDrop(t, drops, 3, "n", false)
	checkDrop(t, drops, 4, "n", false)
}

// TestGraphTopologyNestedBlocks: verifies that an owned variable defined in an
// outer block and used inside an inner block is dropped at the inner block's
// statement index in the outer block (since that's its last-use point).
func TestGraphTopologyNestedBlocks(t *testing.T) {
	// x is owned (Node struct). It is only used inside the if block (index 1).
	// The if block is statement index 1 in main's body.
	// So x's last use (from the outer block's perspective) is at index 1 → dropped there.
	input := `
    package main
    type Node = struct { val: i32 }
	extern fn pass_node(a: #Node)
	fn main() {
		var x = Node{ val: 10 }    // index 0: x is Node (owned)
		if (true) {                // index 1
			pass_node(x)           // block index 0: x used inside if
		}
		pass_node(x)               // index 2: x used again in outer block
	}
	`
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	analyzer := semantic.NewAnalyzer()
	analyzer.Analyze(prog)

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.Solve(prog)

	var mainBlock *ast.BlockStatement
	for _, stmt := range file.Statements {
		if fn, ok := stmt.(*ast.FunctionStatement); ok && fn.Name.Value == "main" {
			mainBlock = fn.Body
			break
		}
	}
	if mainBlock == nil {
		t.Fatal("Could not find main function")
	}

	drops := solver.Drops[mainBlock]
	// fmt.Printf("DEBUG: Nested Drops: %+v\n", drops)

	// x is used inside the if block (index 1) AND at index 2.
	// Its last use in the outer block is at index 2 → dropped after it (3).
	checkDrop(t, drops, 3, "x", true)
}
