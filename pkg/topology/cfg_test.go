package topology_test

import (
	"testing"

	"github.com/nora-language/nora/pkg/lexer"
	"github.com/nora-language/nora/pkg/parser"
	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/semantic"
	"github.com/nora-language/nora/pkg/topology"
)

func TestCFGBuilderSimple(t *testing.T) {
	input := `
	package main
	fn main() {
		var x = 1
		var y = 2
		var z = x + y
	}
	`
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := semantic.NewAnalyzer()
	analyzer.Analyze(prog)

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

	builder := topology.NewCFGBuilder(&analyzer.SemanticInfo)
	start, exit := builder.Build(mainBlock)

	if start == nil {
		t.Fatal("Start node is nil")
	}
	if exit == nil {
		t.Fatal("Exit node is nil")
	}

	// Verify linear control flow structure
	// start (var x) -> node 2 (var y) -> node 3 (var z) -> exit
	node := start
	if node.Kind != topology.BlockSimple {
		t.Errorf("Expected simple node, got kind %v", node.Kind)
	}
	if len(node.Succs) != 1 {
		t.Fatalf("Expected 1 successor, got %d", len(node.Succs))
	}

	node2 := node.Succs[0]
	if len(node2.Succs) != 1 {
		t.Fatalf("Expected 1 successor, got %d", len(node2.Succs))
	}

	node3 := node2.Succs[0]
	if len(node3.Succs) != 1 {
		t.Fatalf("Expected 1 successor, got %d", len(node3.Succs))
	}

	if node3.Succs[0] != exit {
		t.Errorf("Expected z to lead to exit node")
	}
}

func TestCFGBuilderIfElse(t *testing.T) {
	input := `
	package main
	fn main() {
		var x = 1
		if (x > 0) {
			var y = 2
		} else {
			var z = 3
		}
		var final = 4
	}
	`
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := semantic.NewAnalyzer()
	analyzer.Analyze(prog)

	var mainBlock *ast.BlockStatement
	for _, stmt := range file.Statements {
		if fn, ok := stmt.(*ast.FunctionStatement); ok && fn.Name.Value == "main" {
			mainBlock = fn.Body
			break
		}
	}

	builder := topology.NewCFGBuilder(&analyzer.SemanticInfo)
	start, exit := builder.Build(mainBlock)

	if start == nil || exit == nil {
		t.Fatal("CFG nodes are nil")
	}

	// Start (var x) -> Condition (if x > 0)
	condNode := start.Succs[0]
	if condNode.Kind != topology.BlockBranchCondition {
		t.Fatalf("Expected BlockBranchCondition, got %v", condNode.Kind)
	}

	// Condition splits into 2 successors (Then branch & Else branch)
	if len(condNode.Succs) != 2 {
		t.Fatalf("Expected 2 branches, got %d", len(condNode.Succs))
	}

	thenBranch := condNode.Succs[0]
	elseBranch := condNode.Succs[1]

	// Both branch ends must lead to the Join node
	if len(thenBranch.Succs) != 1 || len(elseBranch.Succs) != 1 {
		t.Fatal("Branch ends do not lead to unique single nodes")
	}

	joinNode := thenBranch.Succs[0]
	if joinNode != elseBranch.Succs[0] {
		t.Error("Branches do not merge at a common Join node")
	}

	if joinNode.Kind != topology.BlockJoin {
		t.Errorf("Expected BlockJoin, got %v", joinNode.Kind)
	}

	// Join node -> VarStatement (final) -> exit
	finalNode := joinNode.Succs[0]
	if len(finalNode.Succs) != 1 || finalNode.Succs[0] != exit {
		t.Error("Final statement path to Exit is incorrect")
	}
}

func TestCFGBuilderWhileLoop(t *testing.T) {
	input := `
	package main
	fn main() {
		var x = 1
		while (x < 10) {
			x = x + 1
			if (x == 5) {
				break
			}
		}
	}
	`
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := semantic.NewAnalyzer()
	analyzer.Analyze(prog)

	var mainBlock *ast.BlockStatement
	for _, stmt := range file.Statements {
		if fn, ok := stmt.(*ast.FunctionStatement); ok && fn.Name.Value == "main" {
			mainBlock = fn.Body
			break
		}
	}

	builder := topology.NewCFGBuilder(&analyzer.SemanticInfo)
	start, exit := builder.Build(mainBlock)

	if start == nil || exit == nil {
		t.Fatal("CFG nodes are nil")
	}

	// start (var x) -> Loop Condition
	loopCond := start.Succs[0]
	if loopCond.Kind != topology.BlockLoopCondition {
		t.Fatalf("Expected BlockLoopCondition, got %v", loopCond.Kind)
	}

	// Loop Condition has 2 successors: Loop Body and Loop Exit (Join)
	if len(loopCond.Succs) != 2 {
		t.Fatalf("Expected 2 successors from loop condition, got %d", len(loopCond.Succs))
	}
}

func TestLivenessAnalysis(t *testing.T) {
	input := `
	package main
	fn main() {
		var x = 1
		var y = x
		var z = y + 2
	}
	`
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := semantic.NewAnalyzer()
	analyzer.Analyze(prog)

	var mainBlock *ast.BlockStatement
	for _, stmt := range file.Statements {
		if fn, ok := stmt.(*ast.FunctionStatement); ok && fn.Name.Value == "main" {
			mainBlock = fn.Body
			break
		}
	}

	builder := topology.NewCFGBuilder(&analyzer.SemanticInfo)
	start, exit := builder.Build(mainBlock)

	// Solve liveness on all nodes
	// Use builder's helper or construct a slice of all nodes
	var allNodes []*topology.CFGNode
	var collect func(node *topology.CFGNode, visited map[int]bool)
	collect = func(node *topology.CFGNode, visited map[int]bool) {
		if node == nil || visited[node.ID] {
			return
		}
		visited[node.ID] = true
		if node.ID != -1 {
			allNodes = append(allNodes, node)
		}
		for _, succ := range node.Succs {
			collect(succ, visited)
		}
	}
	collect(start, make(map[int]bool))

	topology.SolveLiveness(allNodes, exit)

	// Verify liveness sets
	// start (var x = 1) -> node2 (var y = x) -> node3 (var z = y + 2) -> exit
	node1 := start
	node2 := node1.Succs[0]
	node3 := node2.Succs[0]

	// Get x and y symbols from semantic analyzer to verify sets
	var symX, symY *semantic.Symbol
	for _, defSym := range analyzer.SemanticInfo.Defs {
		if defSym.Name == "x" {
			symX = defSym
		} else if defSym.Name == "y" {
			symY = defSym
		}
	}

	if symX == nil || symY == nil {
		t.Fatal("Could not resolve semantic symbols for x or y")
	}

	// 1. After node1 (var x = 1), x should be live because it is used in node2
	if !node1.LiveOut[symX] {
		t.Error("Expected symbol 'x' to be LiveOut from node1")
	}

	// 2. Before node2 (var y = x), x should be LiveIn
	if !node2.LiveIn[symX] {
		t.Error("Expected symbol 'x' to be LiveIn to node2")
	}

	// 3. After node2 (var y = x), y should be live because it is used in node3, but x should be dead
	if !node2.LiveOut[symY] {
		t.Error("Expected symbol 'y' to be LiveOut from node2")
	}
	if node2.LiveOut[symX] {
		t.Error("Expected symbol 'x' to be dead (not LiveOut) after node2")
	}

	// 4. Before node3 (var z = y + 2), y should be LiveIn
	if !node3.LiveIn[symY] {
		t.Error("Expected symbol 'y' to be LiveIn to node3")
	}
}
