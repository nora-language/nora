package topology_test

import (
	"strings"
	"testing"

	"github.com/DwiYI/Project-Nora/pkg/lexer"
	"github.com/DwiYI/Project-Nora/pkg/parser"
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/semantic"
	"github.com/DwiYI/Project-Nora/pkg/topology"
)

func runSolverTest(input string) (*topology.Solver, error) {
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	analyzer := semantic.NewAnalyzer()
	analyzer.Loader = &MockPackageLoader{}
	analyzer.Analyze(prog)

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.Solve(prog)
	return solver, nil
}

func assertErrorContains(t *testing.T, solver *topology.Solver, expected string) {
	t.Helper()
	errors := solver.Diagnostics.ErrorMessages()
	found := false
	for _, e := range errors {
		if strings.Contains(e, expected) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error containing '%s', but got: %v", expected, errors)
	}
}

func TestNegUseAfterMove(t *testing.T) {
	input := `
    package main
    type Node = struct { val: i32 }
    fn consume(n: @Node) {}
    fn main() {
        var x = Node{ val: 1 }
        consume(x)
        var y = x // ERROR
    }
    `
	solver, _ := runSolverTest(input)
	assertErrorContains(t, solver, "use of moved value 'x'")
}

func TestNegDoubleMove(t *testing.T) {
	input := `
    package main
    type Node = struct { val: i32 }
    fn consume(n: @Node) {}
    fn main() {
        var x = Node{ val: 1 }
        consume(@x)
        consume(@x) // ERROR
    }
    `
	solver, _ := runSolverTest(input)
	assertErrorContains(t, solver, "use of moved value 'x'")
}

func TestNegAliasUseAfterMove(t *testing.T) {
	input := `
    package main
    type Node = struct { val: i32 }
    fn consume(n: @Node) {}
    fn use_lease(n: #Node) {}
    fn main() {
        var x = Node{ val: 1 }
        var y = #x
        consume(x)
        use_lease(y) // ERROR: origin x was moved
    }
    `
	solver, _ := runSolverTest(input)
	assertErrorContains(t, solver, "use of borrow 'y' whose origin 'x' was moved")
}

func TestNegFieldMove(t *testing.T) {
	input := `
    package main
    type Node = struct { val: i32 }
    fn consume_int(i: @i32) {}
    fn main() {
        var x = Node{ val: 1 }
        consume_int(@x.val)
        var y = x // ERROR: x.val is moved
    }
    `
	// Note: Currently Nora solver tracks field moves but might not invalidate parent struct usage yet.
	// If it doesn't, this test will fail, which is GOOD because it identifies a gap.
	solver, _ := runSolverTest(input)
	assertErrorContains(t, solver, "use of partially moved value 'x'")
}

func TestNegBranchMove(t *testing.T) {
	input := `
    package main
    type Node = struct { val: i32 }
    fn consume(n: @Node) {}
    fn main() {
        var x = Node{ val: 1 }
        if (1 == 1) {
            consume(x)
        }
        var y = x // ERROR: potentially moved
    }
    `
	solver, _ := runSolverTest(input)
	assertErrorContains(t, solver, "use of moved value 'x'")
}

func TestNegImplicitMoveOwnedParam(t *testing.T) {
	input := `
    package main
    type Vector[T] = struct { size: i32 }
    fn (v: #Vector[T]) Push[T](val: T) {}
    fn main() {
        var vec = Vector[str]{ size: 0 }
        var s = "hello"
        vec.Push(s)
        var x = s // ERROR
    }
    `
	solver, _ := runSolverTest(input)
	assertErrorContains(t, solver, "use of moved value 's'")
}

func TestNegFieldMoveInLoop(t *testing.T) {
	input := `
    package main
    type Node = struct { val: i32 }
    fn consume_int(i: @i32) {}
    fn main() {
        var x = Node{ val: 1 }
        var cond = true
        while (cond) {
            consume_int(@x.val)
            cond = false
        }
        var y = x // ERROR: x.val is moved in loop!
    }
    `
	solver, _ := runSolverTest(input)
	assertErrorContains(t, solver, "use of partially moved value 'x'")
}
