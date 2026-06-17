package semantic

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DwiYI/Project-Nora/pkg/lexer"
	"github.com/DwiYI/Project-Nora/pkg/parser"
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/types"
)

// --- 1. MOCK LOADER ---
// Simulates finding "io" on disk
type MockPackageLoader struct{}

func (m *MockPackageLoader) Load(path string) (*Scope, error) {
	if path == "io" {
		scope := NewScope(nil, ScopePackage)

		// Mock io.Print: fn Print(msg: i32) (Read-only borrow)
		// We explicitly define ParamLeases to avoid the "index out of range" panic
		ioPrintType := &types.FunctionType{
			Params:      []types.NRType{types.I32},
			ParamLeases: []types.LeaseKind{types.LeaseRead}, // <--- Vital for Lease Logic
			Return:      types.Void,
		}
		scope.Define("Print", ioPrintType, SymFunc, nil)

		return scope, nil
	}
	return nil, fmt.Errorf("package not found")
}

func TestSemanticLeaseAnalysis(t *testing.T) {
	fmt.Println("\n=== TEST: Semantic Analysis (Lease & Boundaries) ===")

	input := `
    package main
    import "io"

    type User = struct {
        name : str
    }

    // 1. Implicit Read Borrow (Default)
    fn read(_u : #User) {
        // _u is immutable here
    }

    // 2. Explicit Mutable Borrow (&)
    fn update(_u : &User) {
        // _u is mutable here
    }

    // 3. Explicit Consume / Move (@)
    fn destroy(_u : @User) {
        // _u is owned here, caller loses it
    }

    fn main() {
        var u1 = User{ name : "Alice" }

        // Case A: Read (Valid)
        read(u1)
        io.Print(1)

        // Case B: Mutate (Valid)
        update(u1) 
        io.Print(2)

        // Case C: Consume (Valid - Moves ownership)
        destroy(u1) 
        io.Print(3)
    }
    `

	// --- 1. PARSE ---
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser failed: %v", p.Errors())
	}

	prog := &ast.Program{Files: []*ast.File{file}}

	// --- 2. ANALYZE ---
	analyzer := NewAnalyzer()
	analyzer.Loader = &MockPackageLoader{}
	analyzer.Analyze(prog)

	if analyzer.Diagnostics.HasErrors() {
		for _, err := range analyzer.Diagnostics.Diagnostics {
			t.Errorf("Semantic Error: %v", err)
		}
		t.Fatalf("Analysis failed with %d errors", len(analyzer.Diagnostics.Diagnostics))
	}

	// --- 3. VERIFY FUNCTION SIGNATURES (The Contract) ---

	// Check 'read' (Default Read)
	readSym := resolveGlobal(t, analyzer, "read")
	readFn, ok := readSym.Type.(*types.FunctionType)
	if !ok {
		t.Fatal("read is not a function type")
	}

	if len(readFn.ParamLeases) == 0 || readFn.ParamLeases[0] != types.LeaseRead {
		t.Errorf("'read' param should be LeaseRead (0), got %v", readFn.ParamLeases)
	}

	// Check 'update' (Mutable &)
	updateSym := resolveGlobal(t, analyzer, "update")
	updateFn, ok := updateSym.Type.(*types.FunctionType)
	if !ok {
		t.Fatal("update is not a function type")
	}

	if len(updateFn.ParamLeases) == 0 || updateFn.ParamLeases[0] != types.LeaseWrite {
		t.Errorf("'update' param should be LeaseWrite (1), got %v", updateFn.ParamLeases)
	}

	// Check 'destroy' (Consume @)
	destroySym := resolveGlobal(t, analyzer, "destroy")
	destroyFn, ok := destroySym.Type.(*types.FunctionType)
	if !ok {
		t.Fatal("destroy is not a function type")
	}

	if len(destroyFn.ParamLeases) == 0 || destroyFn.ParamLeases[0] != types.LeaseMove {
		t.Errorf("'destroy' param should be LeaseMove (2), got %v", destroyFn.ParamLeases)
	}

	fmt.Println("=== LEASE CONTRACT CHECKS PASSED ===")

	// --- 4. VERIFY USE-AFTER-MOVE (Negative Test) ---
	testUseAfterMove(t)
}

// Helper to find symbols in the global scope Defs map
func resolveGlobal(t *testing.T, sa *SemanticAnalyzer, name string) *Symbol {
	for ident, sym := range sa.SemanticInfo.Defs {
		if ident.Value == name {
			return sym
		}
	}
	t.Fatalf("Could not find global symbol '%s'", name)
	return nil
}

// Sub-test for the "Death" logic (Negative Test Case)
func testUseAfterMove(t *testing.T) {
	fmt.Println("\n=== TEST: Semantic Analysis (Use After Move Check) ===")

	input := `
    package main
    type Data = struct { val: i32 }
    fn consume(d: @Data) {}
    
    fn main() {
        var d = Data{val: 10}
        consume(d) // d is moved/dead now
        
        var x = d // ERROR EXPECTED: Use of moved value
    }
    `
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("move_test.nr")
	prog := &ast.Program{Files: []*ast.File{file}}

	analyzer := NewAnalyzer()
	analyzer.Loader = &MockPackageLoader{}
	analyzer.Analyze(prog)

	// We EXPECT errors here
	if !analyzer.Diagnostics.HasErrors() {
		t.Fatal("Expected 'Use after move' error, but got none!")
	}

	foundMoveError := false
	for _, err := range analyzer.Diagnostics.Diagnostics {
		// Check for the error message substring
		if strings.Contains(err.Message, "use of moved value") {
			foundMoveError = true
			break
		}
	}

	if !foundMoveError {
		t.Errorf("Did not find expected 'use of moved value' error. Got errors: %v", analyzer.Diagnostics.ErrorMessages())
	} else {
		fmt.Println("PASS: Detected illegal use of moved variable.")
	}
}

func TestSemanticAnalysis(t *testing.T) {
	fmt.Println("\n=== TEST: Semantic Analysis (Read Lease Case) ===")

	input := `
    package main
    import "io"

    type User = struct {
        id : str,
        name : str
    }

    fn add(a : i32, b : i32) i32 {
        return a + b
    }

    fn main() {
        var a = 20
        var b = 50 
        b = 300      

        var c = 50 + 20
        var userA = User{ id : "230", name : "andi" }

        io.Print(a)
        io.Print(b)
        io.Print(c)
        io.Print(add(20,30))
    }
    `

	// 1. Parse Phase
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser failed with %d errors: %v", len(p.Errors()), p.Errors())
	}

	// 2. Wrap in Program
	prog := &ast.Program{
		Files: []*ast.File{file},
	}

	// 3. Analyze Phase
	analyzer := NewAnalyzer()
	analyzer.Loader = &MockPackageLoader{}
	analyzer.Analyze(prog)

	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Semantic Analysis Failed: %v", analyzer.Diagnostics.ErrorMessages())
	}

	// 4. Find the 'main' function
	var mainFn *ast.FunctionStatement
	for _, stmt := range file.Statements {
		if f, ok := stmt.(*ast.FunctionStatement); ok && f.Name.Value == "main" {
			mainFn = f
			break
		}
	}

	if mainFn == nil {
		t.Fatal("Could not find 'main' function")
	}

	// 5. Verify variable 'b' and its Lease Context
	stmtB := mainFn.Body.Statements[1].(*ast.VarStatement)
	symbolB := analyzer.SemanticInfo.Defs[stmtB.Name]

	if symbolB == nil {
		t.Fatal("Symbol for 'b' not found in Defs")
	}

	// Verify Type & Lease
	if symbolB.Type.Name() != "i32" {
		t.Errorf("Expected type i32, got %s", symbolB.Type.Name())
	}

	// Nora Specific: Verify the Read Lease
	// Assuming your Symbol struct has a Lease field or constant
	if symbolB.LeaseKind != types.LeaseRead {
		t.Errorf("Expected Read lease for 'b', got %d", symbolB.LeaseKind)
	}

	// 6. Verify Struct Instantiation (userA)
	// userA is index 4 in main: var a(0), var b(1), b=300(2), var c(3), var userA(4)
	stmtUser, ok := mainFn.Body.Statements[4].(*ast.VarStatement)
	if !ok {
		t.Fatalf("Statement[4] is not *ast.VarStatement. got=%T", mainFn.Body.Statements[4])
	}

	symbolUser := analyzer.SemanticInfo.Defs[stmtUser.Name]
	if symbolUser.Type.Name() != "User" {
		t.Errorf("Expected 'userA' to be type 'User', got %s", symbolUser.Type.Name())
	}

	// 7. Verify io.Print Calls (Selector Resolution)
	// Let's check the very last statement: io.Print(add(20,30))
	lastStmt := mainFn.Body.Statements[len(mainFn.Body.Statements)-1].(*ast.ExpressionStatement)
	callExp := lastStmt.Expression.(*ast.CallExpression)

	// Check if the 'io' part of 'io.Print' is resolved as a Package
	selector := callExp.Function.(*ast.SelectorExpression)
	pkgIdent := selector.Left.(*ast.Identifier)

	pkgSymbol := analyzer.SemanticInfo.Uses[pkgIdent]
	if pkgSymbol == nil || pkgSymbol.Kind != SymPackage {
		t.Errorf("Expected 'io' to resolve to a Package symbol")
	}

	fmt.Println("=== ALL SEMANTIC CHECKS PASSED ===")
}

func TestSemanticPinSafety(t *testing.T) {
	input := `
    extern fn sqlite3_open(f: str, db: @ptr)
    
    fn main() {
        var my_db = "aho"
        pin my_db 
        sqlite3_open("test.db", my_db) // ERROR! Trying to @move a pinned var
    }
    `

	// 1. Parse Phase
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("main.nr")

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser failed with %d errors: %v", len(p.Errors()), p.Errors())
	}

	// 2. Wrap in Program
	prog := &ast.Program{
		Files: []*ast.File{file},
	}

	analyzer := NewAnalyzer()
	analyzer.Loader = &MockPackageLoader{}
	analyzer.Analyze(prog)

	// Check for "Unknown Type" errors specifically - these shouldn't happen now
	for _, err := range analyzer.Diagnostics.Diagnostics {
		if strings.Contains(err.Message, "unknown type") {
			t.Fatalf("Setup failure: %v", err)
		}
	}

	// Now check for our intended Safety Error
	foundSafetyError := false
	expectedMsg := "currently pinned and cannot be consumed"

	for _, err := range analyzer.Diagnostics.Diagnostics {
		if strings.Contains(err.Message, expectedMsg) {
			foundSafetyError = true
			break
		}
	}

	if !foundSafetyError {
		t.Errorf("Did not find correct pinned safety error. Got errors: %v", analyzer.Diagnostics.ErrorMessages())
	}
}

// --- 3. TEST ERROR CASES ---
func TestSemanticErrors(t *testing.T) {
	fmt.Println("\n=== TEST: Semantic Errors (Failure Case) ===")

	input := `
    fn main() {
        var x = 10 
        x = 20          
        var _y: str = 50 
        unknown()   
		add(1, 2 + 3, "hello")    
    }
    `
	fmt.Println("--- Step 1: Parsing ---")
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("error.nr")

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser failed unexpectedly: %v", p.Errors())
	}
	fmt.Println("   > Parsing successful.")

	prog := &ast.Program{Files: []*ast.File{file}}

	fmt.Println("--- Step 2: Semantic Analysis ---")
	analyzer := NewAnalyzer()
	analyzer.Analyze(prog)

	// We expect 3 errors:
	// 1. Assign to immutable (x = 20)
	// 2. Type mismatch (y: str = 50)
	// 3. Undefined identifier (unknown)
	// Note: "calling non-function" is suppressed by cascading error fix.
	expectedErrors := 3
	fmt.Printf("   > Expected Errors: %d, Actual Errors: %d\n", expectedErrors, len(analyzer.Diagnostics.Diagnostics))

	if len(analyzer.Diagnostics.Diagnostics) != expectedErrors {
		// Print actual errors to debug
		for i, err := range analyzer.Diagnostics.Diagnostics {
			fmt.Printf("   [%d] %s\n", i+1, err)
		}
		t.Errorf("Expected %d errors, got %d", expectedErrors, len(analyzer.Diagnostics.Diagnostics))
	} else {
		fmt.Println("   > Errors matched expectation.")
	}

	fmt.Println("=== TEST PASSED (Errors were caught correctly) ===")
}

func TestControlFlowAndScoping(t *testing.T) {
	fmt.Println("\n=== TEST: Control Flow & Scoping ===")

	input := `
    fn main() {
        var x = 100

        // 1. Test If Expression Type Compatibility
        // Both branches return i32, so 'result' should be i32
        var result = if (x > 50) {
            1
        } else {
            0
        }
        var y = result
        var z = x + 1

        // 2. Test Scope Shadowing
        // Inner 'x' should shadow outer 'x'
        var outer = x
        {
            var x = "inner" // Shadowing with different type
            var inner_val = x
        }
        
        // Outer 'x' should still be integer
        var check = x + 1 
    }
    `
	l := lexer.New(input, "test.nr")

	p := parser.New(l)
	file := p.Parse("scope.nr")

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	fmt.Println("--- Step 2: Semantic Analysis ---")
	analyzer := NewAnalyzer()
	analyzer.Analyze(file)

	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Analysis failed: %v", analyzer.Diagnostics.ErrorMessages())
	}

	// 4. Find the 'main' function
	var mainFn *ast.FunctionStatement
	for _, stmt := range file.Statements {
		if f, ok := stmt.(*ast.FunctionStatement); ok && f.Name.Value == "main" {
			mainFn = f
			break
		}
	}

	if mainFn == nil {
		t.Fatal("Could not find 'main' function")
	}
	block := mainFn.Body

	// 1. Check 'result' type (should be i32)
	// Index 1
	stmtResult := block.Statements[1].(*ast.VarStatement)
	symResult := analyzer.SemanticInfo.Defs[stmtResult.Name]

	if symResult.Type != types.I32 {
		t.Errorf("Expected 'result' to be i32, got %s", symResult.Type.Name())
	}

	// 2. Check Shadowing
	// Index 6: var check = x + 1
	stmtCheck := block.Statements[6].(*ast.VarStatement)
	symCheck := analyzer.SemanticInfo.Defs[stmtCheck.Name]

	if symCheck.Type != types.I32 {
		t.Errorf("Shadowing failed: expected outer usage to be i32, got %s", symCheck.Type.Name())
	}

	fmt.Println("=== TEST PASSED ===")
}

func TestStructFieldAccess(t *testing.T) {
	fmt.Println("\n=== TEST: Struct Field Access ===")

	input := `
    type Player = struct {
        score: i32,
        tag: str
    }

    fn main() {
        var p = Player{ score: 10, tag: "hero" }
        var s = p.score
        var n = p.tag
    }
    `
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("struct_access.nr")

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	analyzer := NewAnalyzer()
	analyzer.Analyze(&ast.Program{Files: []*ast.File{file}})

	if analyzer.Diagnostics.HasErrors() {
		t.Fatalf("Analysis failed: %v", analyzer.Diagnostics.ErrorMessages())
	}

	// Verify types of 's' and 'n'
	// 4. Find the 'main' function
	var mainFn *ast.FunctionStatement
	for _, stmt := range file.Statements {
		if f, ok := stmt.(*ast.FunctionStatement); ok && f.Name.Value == "main" {
			mainFn = f
			break
		}
	}

	if mainFn == nil {
		t.Fatal("Could not find 'main' function")
	}
	block := mainFn.Body

	// var s = p.score
	stmtS := block.Statements[1].(*ast.VarStatement)
	symS := analyzer.SemanticInfo.Defs[stmtS.Name]

	if symS.Type != types.I32 {
		t.Errorf("Field access failed: p.score expected i32, got %s", symS.Type.Name())
	}

	// var n = p.tag
	stmtN := block.Statements[2].(*ast.VarStatement)
	symN := analyzer.SemanticInfo.Defs[stmtN.Name]

	if symN.Type != types.String {
		t.Errorf("Field access failed: p.tag expected str, got %s", symN.Type.Name())
	}

	fmt.Println("=== TEST PASSED ===")
}

func TestReturnErrors(t *testing.T) {
	fmt.Println("\n=== TEST: Return Type Errors ===")

	input := `
    fn _valid() i32 {
        return 100
    }

    fn _invalid() i32 {
        return "not a number" // Error: type mismatch
    }

    fn _void_mismatch() { // returns void
        return 10 // Error: cannot return value from void function
    }
    `
	l := lexer.New(input, "test.nr")
	p := parser.New(l)
	file := p.Parse("return_err.nr")

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	fmt.Printf("   > Statements Parsed: %d\n", len(file.Statements))
	analyzer := NewAnalyzer()
	analyzer.Analyze(&ast.Program{Files: []*ast.File{file}})

	expectedErrors := 2
	fmt.Printf("   > Expected Errors: %d, Actual: %d\n", expectedErrors, len(analyzer.Diagnostics.Diagnostics))

	if len(analyzer.Diagnostics.Diagnostics) != expectedErrors {
		fmt.Printf("   > Full Parser Output: %v\n", p.Errors())
		for i, err := range analyzer.Diagnostics.Diagnostics {
			fmt.Printf("   [%d] %s\n", i+1, err)
		}
		t.Errorf("Expected %d errors, got %d", expectedErrors, len(analyzer.Diagnostics.Diagnostics))
	} else {
		fmt.Println("=== TEST PASSED ===")
	}
}

func TestCollectSymbolsNilCrash(t *testing.T) {
	fmt.Println("\n=== TEST: CollectSymbols Nil Crash Regression ===")

	analyzer := NewAnalyzer()

	// 1. Test real nil (should not crash, handled by node == nil)
	analyzer.CollectSymbols(nil)

	// 2. Test typed nil pointer (The Crash Case)
	// interface{ast.Node}((*ast.FunctionStatement)(nil))
	var typedNil *ast.FunctionStatement = nil
	var node ast.Node = typedNil

	// Before the fix, this would panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CollectSymbols panicked with typed nil: %v", r)
		}
	}()

	analyzer.CollectSymbols(node)

	// 3. Test in Analyze as well
	analyzer.Analyze(node)

	// 4. Test in resolveTypeNode
	sa := NewAnalyzer()
	var typedNilTypeNode *ast.Identifier = nil
	var typeNode ast.TypeNode = typedNilTypeNode
	sa.resolveTypeNode(typeNode)

	fmt.Println("PASS: No panic with typed nil pointers.")
}

func TestGetPackageNameNilCrash(t *testing.T) {
	fmt.Println("\n=== TEST: GetPackageName Nil Crash Regression ===")
	analyzer := NewAnalyzer()

	// 1. File with nil statement
	file := &ast.File{
		Statements: []ast.Statement{nil},
	}
	name := analyzer.GetPackageName(file)
	if name != "main" {
		t.Errorf("Expected 'main', got %s", name)
	}

	// 2. File with typed nil statement
	var typedNil *ast.PackageStatement = nil
	file.Statements = []ast.Statement{typedNil}
	name = analyzer.GetPackageName(file)
	if name != "main" {
		t.Errorf("Expected 'main', got %s", name)
	}

	// 3. Package statement with nil name
	pkgStmt := &ast.PackageStatement{Name: nil}
	file.Statements = []ast.Statement{pkgStmt}
	name = analyzer.GetPackageName(file)
	if name != "main" {
		t.Errorf("Expected 'main', got %s", name)
	}

	fmt.Println("PASS: GetPackageName is nil-safe.")
}

func TestRecursiveLeaseTypes(t *testing.T) {
	fmt.Println("\n=== TEST: Recursive Lease Types & Implicit Borrows ===")

	input := `
    package main

    type Node = struct {
        val: i32,
        next: #Node
    }

    fn process(_n: #Node) {
        // ...
    }

    fn main() {
        // 1. Test Recursive Struct Creation
        var n1 = Node{ val: 10, next: none }
        var n2 = Node{ val: 20, next: #n1 }

        // 2. Test Implicit Borrow (Owned -> #Lease)
        process(n2) // Should be VALID (Implicit Borrow)

        // 3. Test Implicit Borrow in Struct Literal
        var _n3 = Node{ val: 30, next: n2 } // Should be VALID (Implicit Borrow)
    }
    `

	l := lexer.New(input, "ownership.nr")
	p := parser.New(l)
	file := p.Parse("ownership.nr")

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser failed: %v", p.Errors())
	}

	prog := &ast.Program{Files: []*ast.File{file}}
	analyzer := NewAnalyzer()
	analyzer.Loader = &MockPackageLoader{}
	analyzer.Analyze(prog)

	if analyzer.Diagnostics.HasErrors() {
		for _, err := range analyzer.Diagnostics.Diagnostics {
			t.Errorf("Semantic Error: %v", err)
		}
		t.Fatalf("Analysis failed with %d errors", len(analyzer.Diagnostics.Diagnostics))
	}

	// Verify types
	mainScope := analyzer.GetPackageScope("main")
	nodeSym, ok := mainScope.Resolve("Node")
	if !ok {
		t.Fatal("Could not find 'Node' in main scope")
	}
	nodeType, ok := nodeSym.Type.(*types.StructType)
	if !ok {
		t.Fatal("'Node' is not a struct type")
	}
	nextField := nodeType.Fields["next"]

	if nextField.Name() != "#Node" {
		t.Errorf("Expected field 'next' to have name '#Node', got '%s'", nextField.Name())
	}
	if !nextField.IsLeased() {
		t.Error("Expected field 'next' (#Node) to be marked as Leased")
	}

	fmt.Println("PASS: Recursive leases and implicit borrows verified.")
}
