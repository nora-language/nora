package codegen_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DwiYI/Project-Nora/pkg/codegen"
	"github.com/DwiYI/Project-Nora/pkg/lexer"
	"github.com/DwiYI/Project-Nora/pkg/parser"
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/semantic"
	"github.com/DwiYI/Project-Nora/pkg/topology"
)

// Mock Loader to satisfy Semantic Analyzer
type MockLoader struct{}

func (m *MockLoader) Load(path string) (*semantic.Scope, error) {
	return semantic.NewScope(nil, semantic.ScopePackage), nil
}

func TestEndToEndGeneration(t *testing.T) {
	// 1. Nora SOURCE CODE
	input := `
    fn add(a: &str, b: #str) {
        // Body doesn't matter for signature test
    }

    fn main() {
        var x = "hello"
        var y = "world"
        add(x, y) 
        // 'x' is passed as Mutable (#), so it should get &x
        // 'y' is passed as Read, so just y
        // 'x' and 'y' die here, but since they are stack primitives (int), 
        // the generator should emit comments, not free().
    }
    `

	// 2. PARSE
	l := lexer.New(input, "main.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser Errors: %v", p.Errors())
	}

	// 3. ANALYZE
	analyzer := semantic.NewAnalyzer()
	analyzer.Loader = &MockLoader{}
	analyzer.Analyze(prog)
	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Errors: %v", analyzer.Diagnostics.ErrorMessages())
	}

	// 4. SOLVE TOPOLOGY (Liveness)
	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.Solve(prog)

	// 5. GENERATE C
	gen := codegen.NewGenerator(prog, &analyzer.SemanticInfo, solver, nil, nil)
	code, err := gen.Generate()
	if err != nil {
		t.Fatalf("Codegen Failed: %v", err)
	}

	// 6. VERIFY OUTPUT
	// We check for key C patterns that confirm our logic works.

	expectedSnippets := []string{
		"#include \"runtime.h\"",
		"void add(void* _env_ptr, char** a, char* b)",       // Signature transformation (# -> *)
		"x = ((char*)_str_lit_1.data);", // Var assignment with static literal
		"add(_env_ptr, x, y);",                       // Call desugaring
		"static const struct { nr_header_t h; char data[6]; } _str_lit_1", // Global static variable definition
	}

	t.Logf("Generated C Code:\n%s", code)

	for _, snippet := range expectedSnippets {
		if !strings.Contains(code, snippet) {
			t.Errorf("Expected C output to contain:\n  %q\nBut it was missing.", snippet)
		}
	}
}

func TestDebugLineGeneration(t *testing.T) {
	input := `
    fn test_func() {
        var a = 1
        var b = 2
    }
    `

	l := lexer.New(input, "main.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser Errors: %v", p.Errors())
	}

	analyzer := semantic.NewAnalyzer()
	analyzer.Loader = &MockLoader{}
	analyzer.Analyze(prog)
	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Errors: %v", analyzer.Diagnostics.ErrorMessages())
	}

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.Solve(prog)

	gen := codegen.NewGenerator(prog, &analyzer.SemanticInfo, solver, nil, nil)
	gen.EnableDebug = true
	code, err := gen.Generate()
	if err != nil {
		t.Fatalf("Codegen Failed: %v", err)
	}

	t.Logf("Generated C Code with Debug:\n%s", code)

	// no abs
	// no abs
	expectedLine := fmt.Sprintf("#line 2 \"%s\"\nvoid test_func(void* _env_ptr)", "main.nr")
	// Ensure function signature gets #line directive
	if !strings.Contains(code, expectedLine) {
		t.Errorf("Expected #line directive before function signature, but not found")
	}

	// Ensure no leaked line directives in global main wrappers or auto constructors
	// These are generated at the end, so let's verify that the wrapper function main doesn't have "main.nr" line directive
	wrapperSnippet := "int main(int argc, char** argv) {"
	wrapperIndex := strings.Index(code, wrapperSnippet)
	if wrapperIndex == -1 {
		t.Fatalf("Could not find main wrapper in generated code")
	}

	// Substring after the wrapper
	afterWrapper := code[wrapperIndex:]
	if strings.Contains(afterWrapper, "#line") {
		t.Errorf("Leaked #line preprocessor directives found in trailing wrappers/constructors")
	}
}

func TestSpawnAndParallelLineGeneration(t *testing.T) {
	input := `
    fn worker(x: i32) {
    }

    fn main() {
        spawn worker(42)
        parallel {
            worker(1)
            worker(2)
        }
    }
    `

	l := lexer.New(input, "main.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser Errors: %v", p.Errors())
	}

	analyzer := semantic.NewAnalyzer()
	analyzer.Loader = &MockLoader{}
	analyzer.Analyze(prog)
	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Errors: %v", analyzer.Diagnostics.ErrorMessages())
	}

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.Solve(prog)

	gen := codegen.NewGenerator(prog, &analyzer.SemanticInfo, solver, nil, nil)
	gen.EnableDebug = true
	code, err := gen.Generate()
	if err != nil {
		t.Fatalf("Codegen Failed: %v", err)
	}

	t.Logf("Generated C Code with Spawn/Parallel:\n%s", code)

	// no abs
	// no abs

	// Ensure spawn wrapper signature and body get correct #line mapping to the spawn expression (line 6)
	expectedSpawnWrapper := fmt.Sprintf(`#line 6 "%s"
static void __spawn_wrapper_1(void* p) {
    if (nr_fiber_current() == __nora_step_target_fiber) {
        __nora_step_target_fiber = NULL;
        NR_DEBUGBREAK();
    }
    __nora_fiber_started(nr_fiber_parent());
#line 6 "%s"
    struct __spawn_args_1* args = (struct __spawn_args_1*)p;`, "main.nr", "main.nr")

	if !strings.Contains(code, expectedSpawnWrapper) {
		t.Errorf("Expected #line directive before/inside spawn wrapper, but not found")
	}

	// Ensure parallel wrapper signature and body get correct #line mapping to its statements (lines 8 & 9)
	expectedParallelWrapper1 := fmt.Sprintf(`#line 8 "%s"
static void __parallel_wrapper_2(void* p) {
    if (nr_fiber_current() == __nora_step_target_fiber) {
        __nora_step_target_fiber = NULL;
        NR_DEBUGBREAK();
    }
    __nora_fiber_started(nr_fiber_parent());
#line 8 "%s"
    channel_t* _wg = (channel_t*)p;`, "main.nr", "main.nr")

	expectedParallelWrapper2 := fmt.Sprintf(`#line 9 "%s"
static void __parallel_wrapper_3(void* p) {
    if (nr_fiber_current() == __nora_step_target_fiber) {
        __nora_step_target_fiber = NULL;
        NR_DEBUGBREAK();
    }
    __nora_fiber_started(nr_fiber_parent());
#line 9 "%s"
    channel_t* _wg = (channel_t*)p;`, "main.nr", "main.nr")

	if !strings.Contains(code, expectedParallelWrapper1) {
		t.Errorf("Expected #line directive before/inside parallel wrapper 1, but not found")
	}
	if !strings.Contains(code, expectedParallelWrapper2) {
		t.Errorf("Expected #line directive before/inside parallel wrapper 2, but not found")
	}
}

func TestSelectLineGeneration(t *testing.T) {
	input := `
    fn main() {
        var c1 = make(chan[i32], 1)
        var c2 = make(chan[i32], 1)
        select {
            case var x = <-c1:
                var y = x
            case c2 <- 42:
                var z = 1
            default:
                var w = 0
        }
    }
    `

	l := lexer.New(input, "main.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser Errors: %v", p.Errors())
	}

	analyzer := semantic.NewAnalyzer()
	analyzer.Loader = &MockLoader{}
	analyzer.Analyze(prog)
	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Errors: %v", analyzer.Diagnostics.ErrorMessages())
	}

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.Solve(prog)

	gen := codegen.NewGenerator(prog, &analyzer.SemanticInfo, solver, nil, nil)
	gen.EnableDebug = true
	code, err := gen.Generate()
	if err != nil {
		t.Fatalf("Codegen Failed: %v", err)
	}

	t.Logf("Generated C Code with Select:\n%s", code)

	// Ensure each case branch gets mapped inside the channel select branches correctly
	expectedSelectLines := []string{
		"if (_res == 0)",
		"else if (_res == 1)",
		"else",
	}

	for _, lineSnippet := range expectedSelectLines {
		if !strings.Contains(code, lineSnippet) {
			t.Errorf("Expected select debug snippet not found in output:\n  %s", lineSnippet)
		}
	}
}
