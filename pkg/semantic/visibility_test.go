package semantic

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nora-language/nora/pkg/lexer"
	"github.com/nora-language/nora/pkg/parser"
	"github.com/nora-language/nora/pkg/parser/ast"
)

// Mock loader for visibility tests
type VisibilityMockLoader struct {
	Files map[string]string
}

func (m *VisibilityMockLoader) Load(path string) (*Scope, error) {
	if input, ok := m.Files[path]; ok {
		l := lexer.New(input, path+".nr")
		p := parser.New(l)
		file := p.Parse(path + ".nr")

		sa := NewAnalyzer()
		sa.Loader = m // Allow recursive loading if needed
		sa.CollectSymbols(file)

		return sa.GetPackageScope(sa.GetPackageName(file)), nil
	}
	return nil, fmt.Errorf("package %s not found", path)
}

func TestPackageVisibility(t *testing.T) {
	fmt.Println("\n=== TEST: Package Visibility (pub) ===")

	loader := &VisibilityMockLoader{
		Files: map[string]string{
			"lib": `
				package lib
				
				pub fn PublicFunc() i32 { return 1 }
				fn privateFunc() i32 { return 0 }
				
				pub type PublicStruct = struct { x: i32 }
				type privateStruct = struct { y: i32 }
				
				pub var PublicVar = 10
				var privateVar = 20
			`,
		},
	}

	input := `
		package main
		import "lib"
		
		fn main() {
			var a = lib.PublicFunc()    // Valid
			var b = lib.PublicVar       // Valid
			var c = lib.PublicStruct{x: 1} // Valid
			
			// var d = lib.privateFunc()   // ERROR
			// var e = lib.privateVar      // ERROR
			// var f = lib.privateStruct{y: 2} // ERROR
		}
	`

	l := lexer.New(input, "main.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	analyzer := NewAnalyzer()
	analyzer.Loader = loader
	analyzer.Analyze(prog)

	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Unexpected errors in valid code: %v", analyzer.Diagnostics.ErrorMessages())
	}
	fmt.Println("   > Valid accesses passed.")

	// Test Error Cases
	errorCases := []struct {
		name  string
		input string
		msg   string
	}{
		{
			"private function",
			`package main; import "lib"; fn main() { lib.privateFunc() }`,
			"is private",
		},
		{
			"private variable",
			`package main; import "lib"; fn main() { var x = lib.privateVar }`,
			"is private",
		},
		{
			"private type",
			`package main; import "lib"; fn main() { var x = lib.privateStruct{y:1} }`,
			"is private",
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.input, "test.nr")
			p := parser.New(l)
			file := p.Parse("test.nr")
			prog := &ast.Program{Files: []*ast.File{file}}

			analyzer := NewAnalyzer()
			analyzer.Loader = loader
			analyzer.Analyze(prog)

			if !analyzer.Diagnostics.HasErrors() {
				t.Errorf("Expected error for %s, but got none", tc.name)
			} else {
				found := false
				for _, err := range analyzer.Diagnostics.Diagnostics {
					if strings.Contains(err.Message, tc.msg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message containing %q, got: %v", tc.msg, analyzer.Diagnostics.ErrorMessages())
				}
			}
		})
	}
}
