package topology_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nora-language/nora/pkg/lexer"
	"github.com/nora-language/nora/pkg/parser"
	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/semantic"
	"github.com/nora-language/nora/pkg/topology"
	"github.com/nora-language/nora/pkg/types"
)

type MockPackageLoader struct{}

func (m *MockPackageLoader) Load(path string) (*semantic.Scope, error) {
	if path == "io" {
		scope := semantic.NewScope(nil, semantic.ScopePackage)
		ioPrintType := &types.FunctionType{
			Params:      []types.NRType{types.I32},
			ParamLeases: []types.LeaseKind{types.LeaseRead},
			Return:      types.Void,
		}
		scope.Define("Print", ioPrintType, semantic.SymFunc, nil)
		return scope, nil
	}
	return nil, fmt.Errorf("package not found")
}

// checkDrop verifies whether a variable is (or is not) dropped at a given index.
func checkDrop(t *testing.T, drops map[int][]topology.DropInfo, index int, varName string, shouldExist bool) {
	t.Helper()
	list, ok := drops[index]
	found := false
	if ok {
		for _, info := range list {
			if info.Symbol != nil && info.Symbol.Name == varName {
				found = true
				break
			}
		}
	}
	if shouldExist && !found {
		t.Errorf("Expected '%s' to be dropped after statement %d, but it wasn't. drops=%v", varName, index, drops)
	} else if !shouldExist && found {
		t.Errorf("Did NOT expect '%s' to be dropped after statement %d, but it was.", varName, index)
	}
}

// TestTopologyPinDrop: verifies that `pin` anchors an owned variable's drop to
// the end of the block rather than at its natural last-use point.
func TestTopologyPinDrop(t *testing.T) {
	// x is an owned struct (Node). pin x forces its drop to the last stmt (index 3).
	// pass_node takes #Node (read borrow), so x is NOT consumed at index 2.
	input := `
    package main
    type Node = struct { val: i32 }
	extern fn pass_node(a: #Node)
	fn main() {
		var x = Node{ val: 1 }  // stmt 0
		pin x                    // stmt 1
		pass_node(x)             // stmt 2
		pass_node(x)             // stmt 3
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
	analyzer.Loader = &MockPackageLoader{}
	analyzer.Analyze(prog)

	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Errors: %v", analyzer.Diagnostics.ErrorMessages())
	}

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.DebugMode = true
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
	// fmt.Printf("DEBUG: Final Drops Map: %+v\n", drops)

	// x is pinned → must be dropped at end of block (index 4)
	checkDrop(t, drops, 4, "x", true)
	// x must NOT be dropped early at index 3
	checkDrop(t, drops, 3, "x", false)
}

// TestTopologySolver1: owned struct x is dropped at last read-borrow use;
// owned struct y is moved (consumed) and must NOT be dropped.
func TestTopologySolver1(t *testing.T) {
	input := `
    package main

    type User = struct { name: str }

    fn pass_user(u: #User) {}
    fn consume(u: @User) {}

    fn main() {
        var x = User{ name: "alice" }
        var y = User{ name: "bob" }

        pass_user(x)
        consume(y)
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
	analyzer.Loader = &MockPackageLoader{}
	analyzer.Analyze(prog)

	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Errors: %v", analyzer.Diagnostics.ErrorMessages())
	}

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.DebugMode = true
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

	// Find statement indices dynamically
	idxPassX := -1
	idxConsumeY := -1
	for i, stmt := range mainBlock.Statements {
		if exprStmt, ok := stmt.(*ast.ExpressionStatement); ok {
			if call, ok := exprStmt.Expression.(*ast.CallExpression); ok {
				if ident, ok := call.Function.(*ast.Identifier); ok {
					switch ident.Value {
					case "pass_user":
						idxPassX = i
					case "consume":
						idxConsumeY = i
					}
				}
			}
		}
	}
	if idxPassX == -1 || idxConsumeY == -1 {
		t.Fatalf("Could not locate test statements. pass_user:%d consume:%d", idxPassX, idxConsumeY)
	}

	drops := solver.Drops[mainBlock]
	fmt.Printf("Drops at pass_user(x)[%d] and consume(y)[%d]\n", idxPassX, idxConsumeY)

	// x: owned, last used at pass_user (read borrow) → dropped after it
	checkDrop(t, drops, idxPassX+1, "x", true)
	// y: owned but moved by consume → NOT dropped
	checkDrop(t, drops, idxConsumeY+1, "y", false)
}

// TestTopologySolver2: two owned struct variables; one dropped at last use,
// the other moved (not dropped).
func TestTopologySolver2(t *testing.T) {
	input := `
    package main

    type Item = struct { val: i32 }

    fn pass_item(u: #Item) {}
    fn consume(u: @Item) {}

    fn main() {
        var x = Item{ val: 10 }   // stmt 0
        var y = Item{ val: 20 }   // stmt 1
        pass_item(x)               // stmt 2: last use of x → x dropped here
        consume(y)                 // stmt 3: y moved → y NOT dropped
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
	analyzer.Loader = &MockPackageLoader{}
	analyzer.Analyze(prog)

	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Analysis Failed: %v", analyzer.Diagnostics.ErrorMessages())
	}

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.DebugMode = true
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
	fmt.Printf("Final Drops Map: %+v\n", drops)

	checkDrop(t, drops, 3, "x", true)  // x dropped at last use
	checkDrop(t, drops, 4, "y", false) // y moved, not dropped
}

// TestTopologyAssignment: dependency propagation extends the lifecycle of owned
// providers (a, b) when a borrow-variable (x) that depends on them is used.
func TestTopologyAssignment(t *testing.T) {
	input := `
    package main

    type User = struct { name: str }

    fn pass_user(u: #User) {}

    fn main() {
        var a = User{ name: "alice" }   // Index 0
        var b = User{ name: "bob" }     // Index 1
        var x = #a                       // Index 2: x borrows a
        pass_user(x)                     // Index 3: last use of x (and a)
        x = #b                           // Index 4: x now borrows b
        pass_user(x)                     // Index 5: last use of x (and b)
    }
    `
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	analyzer := semantic.NewAnalyzer()
	analyzer.Loader = &MockPackageLoader{}
	analyzer.Analyze(prog)

	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Errors: %v", analyzer.Diagnostics.ErrorMessages())
	}

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.DebugMode = true
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
	fmt.Printf("Drops Map: %+v\n", drops)

	checkDrop(t, drops, 4, "a", true) // a's lifecycle extends to x's last use at 3
	checkDrop(t, drops, 6, "b", true) // b's lifecycle extends to x's last use at 5
}

// TestTopologyJumps: verifies that owned variables are dropped before a break
// statement exits the containing loop.
func TestTopologyJumps(t *testing.T) {
	input := `
    package main

    type Item = struct { val: i32 }

    fn main() {
        var i = 0
        while (i < 5) {
            var r = Item{ val: 10 }
            if (i == 2) {
                break
            }
            i = i + 1
        }
    }
    `
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	analyzer := semantic.NewAnalyzer()
	analyzer.Loader = &MockPackageLoader{}
	analyzer.Analyze(prog)

	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Errors: %v", analyzer.Diagnostics.ErrorMessages())
	}

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.DebugMode = true
	solver.Solve(prog)

	var ifBody *ast.BlockStatement
	for _, stmt := range file.Statements {
		if fn, ok := stmt.(*ast.FunctionStatement); ok && fn.Name.Value == "main" {
			while := fn.Body.Statements[1].(*ast.WhileStatement)
			whileBody := while.Body
			ifStmt := whileBody.Statements[1].(*ast.ExpressionStatement).Expression.(*ast.IfExpression)
			ifBody = ifStmt.Consequence
			break
		}
	}
	if ifBody == nil {
		t.Fatal("Could not find if body")
	}

	drops := solver.PreDrops[ifBody]
	fmt.Printf("ifBody PreDrops: %+v\n", drops)
	// r is owned and must be dropped at index 0 of ifBody when break is hit
	checkDrop(t, drops, 0, "r", true)
}

// TestTopologyChannels: verifies that sending an owned variable to a channel
// consumes it (move), and receiving from a channel initializes a new owned lifecycle.
func TestTopologyChannels(t *testing.T) {
	input := `
    package main
    type Message = struct { id: i32 }
	fn main() {
		var c = make(chan[Message], 1)
		var m = Message{ id: 1 }
		if (1 == 1) {
			c <- @m                 // m is moved here
		}
		var x = m               // ERROR: use after potential move
		var m2 = <-c            // m2 is a new owned variable
	}
	`
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	analyzer := semantic.NewAnalyzer()
	analyzer.Analyze(prog)

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.DebugMode = true
	solver.Solve(prog)

	// Verify move violation for 'm'
	foundError := false
	errors := solver.Diagnostics.ErrorMessages()
	for _, e := range errors {
		if strings.Contains(e, "use of moved value 'm'") {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected 'use of moved value' error for m in solver diagnostics, but got: %v", errors)
	}

	var mainBlock *ast.BlockStatement
	for _, stmt := range file.Statements {
		if fn, ok := stmt.(*ast.FunctionStatement); ok && fn.Name.Value == "main" {
			mainBlock = fn.Body
			break
		}
	}

	drops := solver.Drops[mainBlock]

	// m2 is owned and used at index 4 (receive), then dropped at the end of block (index 5)
	checkDrop(t, drops, 5, "m2", true)
}

func TestTopologyDefer(t *testing.T) {
	input := `
    package main
    type Message = struct { id: i32 }
	extern fn consume(m: Message)
	fn main() {
		var m = Message{ id: 1 }
		defer consume(m)         // m is captured here
		var x = 1                // some other work
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

	drops := solver.Drops[mainBlock]

	// m should be dropped at the end of the block (index 3), NOT after defer (index 1)
	checkDrop(t, drops, 3, "m", true)
	checkDrop(t, drops, 1, "m", false)
}

func TestTopologyReceiverLeasePropagation(t *testing.T) {
	input := `
    package main
    type Vector[T] = struct { size: i32 }
    fn (v: #Vector[T]) Get[T]() T {
        return "val"
    }
    fn main() {
        var vec = Vector[str]{ size: 1 }
        var x = vec.Get()
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
	// x is propagated as borrow -> must NOT have a drop scheduled!
	checkDrop(t, drops, 2, "x", false)
}

func TestTopologyReassignmentReset(t *testing.T) {
	input := `
    package main
    type Node = struct { val: i32 }
    fn consume_int(i: @i32) {}
    fn main() {
        var x = Node{ val: 1 }
        consume_int(@x.val)
        x = Node{ val: 2 } // Re-assign completely
        var y = x // Should be perfectly legal
    }
    `
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	analyzer := semantic.NewAnalyzer()
	analyzer.Analyze(prog)
	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Errors: %v", analyzer.Diagnostics.ErrorMessages())
	}

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.Solve(prog)
	if solver.Diagnostics.HasErrors() {
		t.Fatalf("Solver Errors: %v", solver.Diagnostics.ErrorMessages())
	}
}

func TestTopologyFieldReassignmentReset(t *testing.T) {
	input := `
    package main
    type Node = struct { val: i32 }
    fn consume_int(i: @i32) {}
    fn main() {
        var x = Node{ val: 1 }
        consume_int(@x.val)
        x.val = 2 // Re-assign field
        var y = x.val // Should be perfectly legal
    }
    `
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	analyzer := semantic.NewAnalyzer()
	analyzer.Analyze(prog)
	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Errors: %v", analyzer.Diagnostics.ErrorMessages())
	}

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.Solve(prog)
	if solver.Diagnostics.HasErrors() {
		t.Fatalf("Solver Errors: %v", solver.Diagnostics.ErrorMessages())
	}
}
