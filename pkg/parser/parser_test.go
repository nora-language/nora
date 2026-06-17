package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nora-language/nora/pkg/lexer"
	"github.com/nora-language/nora/pkg/parser/ast"
)

// DebugPrint updated to handle *ast.File
func DebugPrint(p *Parser, file *ast.File) {
	fmt.Println("--- Parser Errors ---")
	if len(p.Errors()) == 0 {
		fmt.Println("No errors found.")
	} else {
		for _, err := range p.Errors() {
			fmt.Printf("Parser Error: %s\n", err)
		}
	}

	fmt.Println("\n--- AST Structure ---")
	if file != nil {
		fmt.Printf("File Name: %s\n", file.Name)
		fmt.Printf("File Content: %s\n", file.String())
		for i, stmt := range file.Statements {
			fmt.Printf("Statement [%d]: %T\n", i, stmt)
		}
	}
	fmt.Println("----------------------")
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestVar(t *testing.T) {
	input := `var aho : i32 = 20`

	l := lexer.New(input, "test.nr")
	p := New(l)

	file := p.Parse("test.nr")
	DebugPrint(p, file)
	checkParserErrors(t, p)

	if len(file.Statements) != 1 {
		t.Fatalf("file.Statements does not contain 1 statement. got=%d", len(file.Statements))
	}

	stmt, ok := file.Statements[0].(*ast.VarStatement)
	if !ok {
		t.Fatalf("stmt not *ast.VarStatement. got=%T", file.Statements[0])
	}

	// 1. Check Name
	if stmt.Name.Value != "aho" {
		t.Errorf("stmt.Name.Value not 'aho'. got=%s", stmt.Name.Value)
	}

	// 2. Check Type (i32)
	if stmt.Type == nil {
		t.Fatal("stmt.Type is nil")
	}
	if stmt.Type.String() != "i32" {
		t.Errorf("stmt.Type.String() not 'i32'. got=%s", stmt.Type.String())
	}

	// 3. Check Value (20)
	if stmt.Value == nil {
		t.Fatal("stmt.Value is nil")
	}

	// We expect stmt.Value to be an *ast.IntegerLiteral
	val, ok := stmt.Value.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("stmt.Value is not *ast.IntegerLiteral. got=%T", stmt.Value)
	}
	if val.Value != 20 {
		t.Errorf("val.Value not 20. got=%d", val.Value)
	}
}

func TestIfElseExpression(t *testing.T) {
	// 1. input with a condition, a 'then' block, and an 'else' block
	input := `if (x < y) { x } else { y }`

	l := lexer.New(input, "test.nr")
	p := New(l)

	file := p.Parse("test.nr")
	DebugPrint(p, file)
	checkParserErrors(t, p)

	// 2. Validate we have exactly 1 statement
	if len(file.Statements) != 1 {
		t.Fatalf("file.Statements does not contain 1 statement. got=%d", len(file.Statements))
	}

	// 3. Unwrap ExpressionStatement (Standard for expression-based languages)
	// Note: If 'if' is a statement in your AST, cast directly to *ast.IfStatement
	stmt, ok := file.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T", file.Statements[0])
	}

	// 4. Cast the expression to *ast.IfExpression
	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T", stmt.Expression)
	}

	// 5. Test the Condition (x < y)
	// We expect an InfixExpression here
	cond, ok := exp.Condition.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("exp.Condition is not ast.InfixExpression. got=%T", exp.Condition)
	}

	if cond.Left.String() != "x" {
		t.Errorf("condition.Left is not 'x'. got=%s", cond.Left.String())
	}
	if cond.Operator != "<" {
		t.Errorf("condition.Operator is not '<'. got=%s", cond.Operator)
	}
	if cond.Right.String() != "y" {
		t.Errorf("condition.Right is not 'y'. got=%s", cond.Right.String())
	}

	// 6. Test the Consequence (The "If" Block)
	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got=%d", len(exp.Consequence.Statements))
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Consequence statement is not ast.ExpressionStatement. got=%T", exp.Consequence.Statements[0])
	}
	if consequence.Expression.String() != "x" {
		t.Errorf("Consequence expression not 'x'. got=%s", consequence.Expression.String())
	}

	// 7. Test the Alternative (The "Else" Block)
	if exp.Alternative == nil {
		t.Fatal("exp.Alternative is nil. Expected else block")
	}

	altBlock, ok := exp.Alternative.(*ast.BlockStatement)
	if !ok {
		t.Fatalf("Alternative is not ast.BlockStatement. got=%T", exp.Alternative)
	}

	if len(altBlock.Statements) != 1 {
		t.Errorf("alternative is not 1 statements. got=%d", len(altBlock.Statements))
	}

	alternative, ok := altBlock.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Alternative statement is not ast.ExpressionStatement. got=%T", altBlock.Statements[0])
	}
	if alternative.Expression.String() != "y" {
		t.Errorf("Alternative expression not 'y'. got=%s", alternative.Expression.String())
	}
}

func TestIfElseNewLineExpression(t *testing.T) {
	// 1. Simple else on newline
	input1 := `if (x < y) { x }
	else { y }`

	l1 := lexer.New(input1, "test.nr")
	p1 := New(l1)

	file1 := p1.Parse("test.nr")
	DebugPrint(p1, file1)
	checkParserErrors(t, p1)

	// 2. Chain of else-if on newlines
	input2 := `if (x < y) { x }
	else if (x > y) { y }
	else { z }`

	l2 := lexer.New(input2, "test.nr")
	p2 := New(l2)

	file2 := p2.Parse("test.nr")
	DebugPrint(p2, file2)
	checkParserErrors(t, p2)
}

func TestExternFunctionStatement(t *testing.T) {
	input := `
	extern fn printf(s: str)
	extern fn sqlite3_open(f: str, db: @ptr) i32
	`
	l := lexer.New(input, "test.nr")
	p := New(l)
	file := p.Parse("test.nr")
	DebugPrint(p, file)
	checkParserErrors(t, p)

	if len(file.Statements) != 2 {
		t.Fatalf("file.Statements does not contain 2 statements. got=%d",
			len(file.Statements))
	}

	tests := []struct {
		expectedName   string
		expectedParams int
		expectedExtern bool
		hasReturnType  bool
	}{
		{"printf", 1, true, false},
		{"sqlite3_open", 2, true, true},
	}

	for i, tt := range tests {
		stmt, ok := file.Statements[i].(*ast.FunctionStatement)
		if !ok {
			t.Fatalf("stmt[%d] is not *ast.FunctionStatement. got=%T", i, file.Statements[i])
		}

		if stmt.Name.Value != tt.expectedName {
			t.Errorf("stmt[%d] name wrong. expected=%s, got=%s", i, tt.expectedName, stmt.Name.Value)
		}

		if len(stmt.Parameters) != tt.expectedParams {
			t.Errorf("stmt[%d] params wrong. expected=%d, got=%d", i, tt.expectedParams, len(stmt.Parameters))
		}

		if stmt.IsExtern != tt.expectedExtern {
			t.Errorf("stmt[%d] IsExtern wrong. expected=%t, got=%t", i, tt.expectedExtern, stmt.IsExtern)
		}

		if stmt.Body != nil {
			t.Errorf("stmt[%d] body should be nil for extern", i)
		}

		if tt.hasReturnType && stmt.ReturnType == nil {
			t.Errorf("stmt[%d] expected return type but got nil", i)
		}
	}
}

func TestFunction(t *testing.T) {
	// Input has 3 parameters: Implicit Read (a), Explicit Write (#b), Explicit Move (@c)
	input := `fn add(a : i32, b : #i32, c : @i32) i32 {
        return a + b + c
    }`

	l := lexer.New(input, "test.nr")
	p := New(l)

	file := p.Parse("test.nr")
	DebugPrint(p, file)
	checkParserErrors(t, p)

	if len(file.Statements) != 1 {
		t.Fatalf("file.Statements does not contain 1 statement. got=%d", len(file.Statements))
	}

	// 1. Check if it's a Function Statement
	stmt, ok := file.Statements[0].(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("stmt not *ast.FunctionStatement. got=%T", file.Statements[0])
	}

	// 2. Check Name
	if stmt.Name.Value != "add" {
		t.Errorf("function name not 'add'. got=%s", stmt.Name.Value)
	}

	// 3. Check Parameters Count (Updated to 3)
	if len(stmt.Parameters) != 3 {
		t.Fatalf("expected 3 parameters, got=%d", len(stmt.Parameters))
	}

	// 4. Verify Parameter Names and Types
	tests := []struct {
		Index        int
		ExpectedName string
		ExpectedType string
	}{
		{0, "a", "i32"},
		{1, "b", "(#i32)"},
		{2, "c", "(@i32)"},
	}

	for _, tt := range tests {
		param := stmt.Parameters[tt.Index]

		// Check Name
		if param.Name.Value != tt.ExpectedName {
			t.Errorf("param[%d] name wrong. expected=%q, got=%q",
				tt.Index, tt.ExpectedName, param.Name.Value)
		}

		// Check Type
		if param.Type.String() != tt.ExpectedType {
			t.Errorf("param[%d] '%s' type wrong. expected=%s, got=%s",
				tt.Index, tt.ExpectedName, tt.ExpectedType, param.Type.String())
		}
	}

	// 5. Check Return Type
	if stmt.ReturnType.String() != "i32" {
		t.Errorf("return type not 'i32'. got=%s", stmt.ReturnType.String())
	}
}
func TestPinStatement(t *testing.T) {
	input := `
    var path = "data.db"
    pin path, my_db
    `
	l := lexer.New(input, "test.nr")
	p := New(l)

	file := p.Parse("test.nr")
	checkParserErrors(t, p)

	if len(file.Statements) != 2 {
		t.Fatalf("file.Statements does not contain 2 statements. got=%d",
			len(file.Statements))
	}

	pinStmt, ok := file.Statements[1].(*ast.PinStatement)
	if !ok {
		t.Fatalf("stmt 1 is not *ast.PinStatement. got=%T", file.Statements[1])
	}

	if len(pinStmt.Targets) != 2 {
		t.Fatalf("expected 2 pin targets, got=%d", len(pinStmt.Targets))
	}

	if pinStmt.Targets[0].Value != "path" {
		t.Errorf("pin target[0] not 'path'. got=%s", pinStmt.Targets[0].Value)
	}

	if pinStmt.Targets[1].Value != "my_db" {
		t.Errorf("pin target[1] not 'my_db'. got=%s", pinStmt.Targets[1].Value)
	}
}

func TestBooleanLiteral(t *testing.T) {
	tests := []struct {
		input           string
		expectedBoolean bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input, "test.nr")
		p := New(l)

		file := p.Parse("test.nr")

		DebugPrint(p, file)
		checkParserErrors(t, p)

		if len(file.Statements) != 1 {
			t.Fatalf("file has not enough statements. got=%d",
				len(file.Statements))
		}

		exprStmt, ok := file.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("file.Statements[0] is not ast.ExpressionStatement. got=%T",
				file.Statements[0])
		}

		boolean, ok := exprStmt.Expression.(*ast.Boolean)
		if !ok {
			t.Fatalf("exp not *ast.Boolean. got=%T", exprStmt.Expression)
		}

		if boolean.Value != tt.expectedBoolean {
			t.Errorf("boolean.Value not %t. got=%t", tt.expectedBoolean,
				boolean.Value)
		}
	}
}

func TestVarStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"var x = 5", "x", 5},
		{"var x : i32 = 5", "x", 5},
		{"var y = true", "y", true},
		{"var foobar = y", "foobar", "y"},
		{"var foobar = 0xffffff", "foobar", "0xffffff"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input, "test.nr")
		p := New(l)

		file := p.Parse("test.nr")

		DebugPrint(p, file)
		checkParserErrors(t, p)

		if len(file.Statements) != 1 {
			t.Fatalf("file.Statements does not contain 1 statement. got=%d", len(file.Statements))
		}

		stmt := file.Statements[0].(*ast.VarStatement)
		if stmt.Name.Value != tt.expectedIdentifier {
			t.Errorf("stmt.Name.Value not '%s'. got=%s", tt.expectedIdentifier, stmt.Name.Value)
		}
	}
}

func TestCallExpression(t *testing.T) {
	// 1. Updated input to a real function call
	input := `add(1, 2 + 3, "hello")`

	l := lexer.New(input, "test.nr")
	p := New(l)

	file := p.Parse("test.nr")

	DebugPrint(p, file)
	checkParserErrors(t, p)

	if len(file.Statements) != 1 {
		t.Fatalf("file.Statements does not contain 1 statement. got=%d", len(file.Statements))
	}

	// 2. Top-level of a standalone call is an ExpressionStatement
	stmt, ok := file.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not *ast.ExpressionStatement. got=%T", file.Statements[0])
	}

	// 3. The expression inside must be a CallExpression
	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.CallExpression. got=%T", stmt.Expression)
	}

	// 4. Check the Function Name (Identifier)
	ident, ok := exp.Function.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp.Function is not *ast.Identifier. got=%T", exp.Function)
	}
	if ident.Value != "add" {
		t.Errorf("ident.Value not 'add'. got=%s", ident.Value)
	}

	// 5. Check Arguments (using your []*ArgumentsExpression slice)
	if len(exp.Arguments) != 3 {
		t.Fatalf("wrong number of arguments. got=%d", len(exp.Arguments))
	}

	// Verify first argument: 1
	verifyLiteralValue(t, exp.Arguments[0].Value, 1)

	// Verify second argument: 2 + 3 (InfixExpression)
	_, ok = exp.Arguments[1].Value.(*ast.InfixExpression)
	if !ok {
		t.Errorf("second argument is not InfixExpression. got=%T", exp.Arguments[1].Value)
	}

	// Verify third argument: "hello"
	verifyLiteralValue(t, exp.Arguments[2].Value, "hello")
}

// Helper function to keep the test clean
func verifyLiteralValue(t *testing.T, exp ast.Expression, expected interface{}) {
	switch v := expected.(type) {
	case int:
		lit, ok := exp.(*ast.IntegerLiteral)
		if !ok || lit.Value != int64(v) {
			t.Errorf("literal value not %d. got=%v", v, exp)
		}
	case string:
		lit, ok := exp.(*ast.StringLiteral)
		if !ok || lit.Value != v {
			t.Errorf("literal value not %s. got=%v", v, exp)
		}
	}
}

func TestTypeStatements(t *testing.T) {
	input := `
    type User = struct {
        id: i32,
        name: str,
        aho : f64
    }
    `
	l := lexer.New(input, "test.nr")
	p := New(l)

	file := p.Parse("test.nr")

	DebugPrint(p, file)
	checkParserErrors(t, p)

	if len(file.Statements) == 0 {
		t.Fatalf("file.Statements is empty")
	}

	stmt, ok := file.Statements[0].(*ast.TypeStatement)
	if !ok {
		t.Fatalf("stmt is not *ast.TypeStatement. got=%T", file.Statements[0])
	}

	if stmt.Name.Value != "User" {
		t.Fatalf("stmt.Name.Value not 'User'. got=%s", stmt.Name.Value)
	}

	structLit, ok := stmt.Value.(*ast.StructLiteral)
	if !ok {
		t.Fatalf("stmt.Value is not *ast.StructLiteral. got=%T", stmt.Value)
	}

	if len(structLit.Fields) != 3 {
		t.Errorf("structLit.Fields length not 3. got=%d", len(structLit.Fields))
	}
}

func TestSpawnExpression(t *testing.T) {
	input := `
    spawn worker(1, 2)
    var task = spawn compute()
    `
	l := lexer.New(input, "test.nr")
	p := New(l)

	file := p.Parse("test.nr")

	DebugPrint(p, file)
	checkParserErrors(t, p)

	if len(file.Statements) != 2 {
		t.Fatalf("file.Statements does not contain 2 statements. got=%d",
			len(file.Statements))
	}

	// Test 1: spawn worker(1, 2)
	exprStmt, ok := file.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt 0 is not ExpressionStatement. got=%T", file.Statements[0])
	}
	_, ok = exprStmt.Expression.(*ast.SpawnExpression)
	if !ok {
		t.Errorf("expr 0 is not SpawnExpression. got=%T", exprStmt.Expression)
	}

	// Test 2: var task = spawn compute()
	varStmt, ok := file.Statements[1].(*ast.VarStatement)
	if !ok {
		t.Fatalf("stmt 1 is not VarStatement. got=%T", file.Statements[1])
	}
	if varStmt.Name.Value != "task" {
		t.Errorf("varStmt name not 'task'. got=%s", varStmt.Name.Value)
	}
}

func TestMatchExpression(t *testing.T) {
	input := `
    match result {
        Ok(val) => val,
        Err(e) => none
    }
    `
	l := lexer.New(input, "test.nr")
	p := New(l)

	file := p.Parse("test.nr")

	DebugPrint(p, file)
	checkParserErrors(t, p)

	stmt := file.Statements[0].(*ast.ExpressionStatement)
	_, ok := stmt.Expression.(*ast.MatchExpression)
	if !ok {
		t.Errorf("stmt.Expression is not MatchExpression. got=%T", stmt.Expression)
	}
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b", "((-a) * b)"},
		{"a + b + c", "((a + b) + c)"},
		{"a + b * c", "(a + (b * c))"},
		{"a * b / c", "((a * b) / c)"},
		{"5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))"},
		{"spawn worker(1 + 2)", "spawn worker((1 + 2))"},
		{"arr[0].id?", "((arr[0]).id)?"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input, "test.nr")
		p := New(l)

		file := p.Parse("test.nr")

		DebugPrint(p, file)
		checkParserErrors(t, p)

		// TrimSpace is needed because ast.File.String() adds newlines
		actual := strings.TrimSpace(file.String())
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func TestParserPackageEnforcement(t *testing.T) {
	input := `var x : i32 = 10`

	// Case 1: AllowNoPackage = true (should succeed)
	{
		l := lexer.New(input, "foo.nr")
		p := New(l)
		p.AllowNoPackage = true
		_ = p.Parse("foo.nr")
		if len(p.Errors()) > 0 {
			t.Errorf("expected no errors when AllowNoPackage is true, got: %v", p.Errors())
		}
	}

	// Case 2: AllowNoPackage = false (should report expected error)
	{
		l := lexer.New(input, "foo.nr")
		p := New(l)
		p.AllowNoPackage = false
		_ = p.Parse("foo.nr")
		errors := p.Errors()
		if len(errors) == 0 {
			t.Errorf("expected error when AllowNoPackage is false and file has no package")
		} else {
			expectedErr := "expected package statement at the beginning of the file"
			found := false
			for _, errStr := range errors {
				if strings.Contains(errStr, expectedErr) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error containing %q, got errors: %v", expectedErr, errors)
			}
		}
	}

	// Case 3: AllowNoPackage = false, but file has package statement (should succeed)
	{
		inputWithPkg := "package main\nvar x : i32 = 10"
		l := lexer.New(inputWithPkg, "foo.nr")
		p := New(l)
		p.AllowNoPackage = false
		_ = p.Parse("foo.nr")
		if len(p.Errors()) > 0 {
			t.Errorf("expected no errors when AllowNoPackage is false and file has package statement, got: %v", p.Errors())
		}
	}
}
