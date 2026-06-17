package lsp

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestURIToPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		uri := "file:///C:/Users/test/project/file.nr"
		want := `c:\users\test\project\file.nr`
		got := uriToPath(uri)
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	} else {
		uri := "file:///home/test/project/file.nr"
		want := "/home/test/project/file.nr"
		got := uriToPath(uri)
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	}
}

func TestLSPBasic(t *testing.T) {
	handler := NewHandler()

	wd, _ := os.Getwd()
	// wd is e:\Project\Project Nora\second\pkg\lsp
	// we want the root e:\Project\Project Nora\second
	root := filepath.Dir(filepath.Dir(wd))

	uri := "file:///" + filepath.ToSlash(filepath.Join(root, "test.nr"))
	content := `
	package main
	
	fn add(a: i32, b: i32) i32 {
		return a + b
	}

	fn main() {
		var x = 10
		var y = add(x, 20)
	}
	`

	// Test DidOpen
	err := handler.TextDocumentDidOpen(context.Background(), nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  uri,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Verify document loaded
	d, ok := handler.docs.Load(uri)
	if !ok {
		t.Fatalf("Document not loaded")
	}
	doc := d.(*Document)
	if doc.Program == nil || doc.Info == nil {
		t.Fatalf("Analysis failed")
	}

	if len(doc.Diags.Diagnostics) > 0 {
		for _, d := range doc.Diags.Diagnostics {
			t.Logf("DIAG: %s", d.Message)
		}
	}

	// Test Hover over 'add' in 'add(x, 20)'
	// In the log, 'add' call was at {10 11 0} (Line 10, Col 11)
	hover, err := handler.TextDocumentHover(context.Background(), nil, &HoverParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 9, Character: 11},
	})
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}
	if hover != nil {
		t.Logf("Hover: %v", hover.Contents.Value)
	}

	// Test Completion
	items, err := handler.TextDocumentCompletion(context.Background(), nil, &CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 10, Character: 2},
	})
	if err != nil {
		t.Fatalf("Completion failed: %v", err)
	}

	foundAdd := false
	for _, item := range items {
		if item.Label == "add" {
			foundAdd = true
			break
		}
	}

	if !foundAdd {
		t.Errorf("Expected completion to contain 'add', got %v", items)
	}

	// Test Dot Completion
	uri2 := "file:///" + filepath.ToSlash(filepath.Join(root, "test2.nr"))
	content2 := "package main\nimport \"io\"\nfn main() {\n    io.\n}\n"
	handler.TextDocumentDidOpen(context.Background(), nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri2, Text: content2},
	})

	items2, err := handler.TextDocumentCompletion(context.Background(), nil, &CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: uri2},
		Position:     Position{Line: 3, Character: 7},
	})
	if err != nil {
		t.Fatalf("Dot completion failed: %v", err)
	}

	t.Logf("Dot completion items: %v", items2)
	if docVal, ok := handler.docs.Load(uri2); ok {
		d := docVal.(*Document)
		for _, diag := range d.Diags.Diagnostics {
			t.Logf("[DEBUG TestLSPBasic] DIAG on test2.nr: %s", diag.Message)
		}
	}
	foundPrint := false
	for _, item := range items2 {
		if item.Label == "PrintLn" || item.Label == "Print" {
			foundPrint = true
			break
		}
	}
	if !foundPrint {
		t.Errorf("Expected dot completion to contain 'PrintLn', got %v", items2)
	}

	// Test Hover Signature
	// 'fn add' is at line 4, char 4 in the log it was {4 5 0} -> Line 4, Col 5
	hover, err = handler.TextDocumentHover(context.Background(), nil, &HoverParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 3, Character: 4},
	})
	if err != nil || hover == nil {
		t.Fatalf("Hover failed: %v", err)
	}
	if !strings.Contains(hover.Contents.Value, "fn(i32, i32) i32") {
		t.Errorf("Expected hover to contain signature, got %q", hover.Contents.Value)
	}

	// Test Go to Definition for 'add'
	def, err := handler.TextDocumentDefinition(context.Background(), nil, &DefinitionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 9, Character: 11}, // add(x, 20)
	})
	if err != nil || def == nil {
		t.Fatalf("Definition failed: %v", err)
	}
	t.Logf("Definition URI: %s", def.URI)
	if !strings.HasSuffix(def.URI, "test.nr") {
		t.Errorf("Expected definition URI to end with test.nr, got %s", def.URI)
	}
	if def.Range.Start.Line != 3 { // fn add is at line 4 (index 3)
		t.Errorf("Expected definition at line 3, got %d", def.Range.Start.Line)
	}

	// Test Go to Definition for 'println' in 'io.println'
	// uri2 is "package main\nimport \"io\"\nfn main() {\n    io.\n}\n"
	// Let's change content2 to have 'io.PrintLn("hello")'
	content3 := "package main\nimport \"io\"\nfn main() {\n    io.PrintLn(\"hello\")\n}\n"
	uri3 := "file:///" + filepath.ToSlash(filepath.Join(root, "test3.nr"))
	handler.TextDocumentDidOpen(context.Background(), nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri3, Text: content3},
	})

	def3, err := handler.TextDocumentDefinition(context.Background(), nil, &DefinitionParams{
		TextDocument: TextDocumentIdentifier{URI: uri3},
		Position:     Position{Line: 3, Character: 10}, // io.PrintLn
	})
	if err != nil || def3 == nil {
		t.Fatalf("Cross-file Definition failed: %v", err)
	}
	t.Logf("Cross-file Definition URI: %s", def3.URI)
	if !strings.HasSuffix(def3.URI, "io.nr") {
		t.Errorf("Expected definition URI to end with io.nr, got %s", def3.URI)
	}

	// Test Go to Definition for the import path "io"
	def4, err := handler.TextDocumentDefinition(context.Background(), nil, &DefinitionParams{
		TextDocument: TextDocumentIdentifier{URI: uri3},
		Position:     Position{Line: 1, Character: 9}, // import "io"
	})
	if err != nil || def4 == nil {
		t.Fatalf("Import path Definition failed: %v", err)
	}
	t.Logf("Import path Definition URI: %s", def4.URI)
	// Points to the actual library file!
	if !strings.HasSuffix(def4.URI, "io.nr") {
		t.Errorf("Expected definition URI to end with io.nr, got %s", def4.URI)
	}
}

func TestLSPInlayHintAndNilPanic(t *testing.T) {
	handler := NewHandler()

	wd, _ := os.Getwd()
	root := filepath.Dir(filepath.Dir(wd))
	uri := "file:///" + filepath.ToSlash(filepath.Join(root, "test_nil_panic.nr"))

	content := `
	package main
	type User = struct {
		name: str
	}
	fn main() {
		var u = User{ name: "Alice" }
	}
	`

	err := handler.TextDocumentDidOpen(context.Background(), nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  uri,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Verify that Inlay Hint is generated and has no newlines
	hints, err := handler.TextDocumentInlayHint(context.Background(), nil, &InlayHintParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Range: Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: 10, Character: 0},
		},
	})
	if err != nil {
		t.Fatalf("Inlay Hint failed: %v", err)
	}

	foundHint := false
	for _, hint := range hints {
		if strings.Contains(hint.Label, "struct User") {
			foundHint = true
			if strings.Contains(hint.Label, "\n") || strings.Contains(hint.Label, "\r") {
				t.Errorf("Inlay hint contains newlines: %q", hint.Label)
			}
		}
	}
	if !foundHint {
		t.Logf("Hints: %v", hints)
	}

	// Test nodeMaxPosition typed nil safety by verifying Document Symbols doesn't panic
	_, err = handler.TextDocumentDocumentSymbol(context.Background(), nil, &DocumentSymbolParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("Document symbols failed: %v", err)
	}
}

func TestLSPNetNRCrash(t *testing.T) {
	handler := NewHandler()

	wd, _ := os.Getwd()
	root := filepath.Dir(filepath.Dir(wd))

	netPath := filepath.Join(root, "std", "net", "net.nr")
	bytes, err := os.ReadFile(netPath)
	if err != nil {
		t.Fatalf("Failed to read std/net/net.nr: %v", err)
	}
	content := string(bytes)
	uri := "file:///" + filepath.ToSlash(netPath)

	err = handler.TextDocumentDidOpen(context.Background(), nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  uri,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen net.nr failed: %v", err)
	}

	_, err = handler.TextDocumentInlayHint(context.Background(), nil, &InlayHintParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Range: Range{
			Start: Position{Line: 0, Character: 0},
			End:   Position{Line: 300, Character: 0},
		},
	})
	if err != nil {
		t.Logf("Inlay Hint returned error: %v", err)
	}

	_, err = handler.TextDocumentDocumentSymbol(context.Background(), nil, &DocumentSymbolParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Logf("Document Symbol returned error: %v", err)
	}
}

func TestLSPJsonDiagnostics(t *testing.T) {
	handler := NewHandler()

	wd, _ := os.Getwd()
	root := filepath.Dir(filepath.Dir(wd))

	jsonPath := filepath.Join(root, "std", "json", "json.nr")
	jsonBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("Failed to read std/json/json.nr: %v", err)
	}
	jsonURI := "file:///" + filepath.ToSlash(jsonPath)

	netPath := filepath.Join(root, "std", "net", "net.nr")
	netBytes, err := os.ReadFile(netPath)
	if err != nil {
		t.Fatalf("Failed to read std/net/net.nr: %v", err)
	}
	netURI := "file:///" + filepath.ToSlash(netPath)

	// Open json.nr first
	err = handler.TextDocumentDidOpen(context.Background(), nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  jsonURI,
			Text: string(jsonBytes),
		},
	})
	if err != nil {
		t.Fatalf("DidOpen json.nr failed: %v", err)
	}

	d, ok := handler.docs.Load(jsonURI)
	if !ok {
		t.Fatalf("json.nr document not loaded")
	}
	docJson := d.(*Document)
	if len(docJson.Diags.Diagnostics) > 0 {
		for _, diag := range docJson.Diags.Diagnostics {
			t.Errorf("First open JSON DIAG: %s (line %d)", diag.Message, diag.Range.Start.Line)
		}
	}

	// Open net.nr
	err = handler.TextDocumentDidOpen(context.Background(), nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  netURI,
			Text: string(netBytes),
		},
	})
	if err != nil {
		t.Fatalf("DidOpen net.nr failed: %v", err)
	}

	d2, ok := handler.docs.Load(netURI)
	if !ok {
		t.Fatalf("net.nr document not loaded")
	}
	docNet := d2.(*Document)
	if len(docNet.Diags.Diagnostics) > 0 {
		for _, diag := range docNet.Diags.Diagnostics {
			t.Logf("Net DIAG: %s (line %d)", diag.Message, diag.Range.Start.Line)
		}
	}

	// Open/update json.nr again
	err = handler.TextDocumentDidOpen(context.Background(), nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  jsonURI,
			Text: string(jsonBytes),
		},
	})
	if err != nil {
		t.Fatalf("Re-open json.nr failed: %v", err)
	}

	d3, ok := handler.docs.Load(jsonURI)
	if !ok {
		t.Fatalf("json.nr document not loaded on second open")
	}
	docJson3 := d3.(*Document)
	if len(docJson3.Diags.Diagnostics) > 0 {
		for _, diag := range docJson3.Diags.Diagnostics {
			t.Errorf("Second open JSON DIAG: %s (line %d)", diag.Message, diag.Range.Start.Line)
		}
	}
}

func TestSemanticTokensReceiver(t *testing.T) {
	handler := NewHandler()
	ctx := context.Background()

	uri := "file:///test.nr"
	content := `package main
type Point = struct {
    x: i32
}
fn (p: #Point) GetX() Point {
    return p
}`
	err := handler.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  uri,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	tokens, err := handler.TextDocumentSemanticTokensFull(ctx, nil, &SemanticTokensParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("Semantic tokens failed: %v", err)
	}

	// Legend Token Types for reference:
	// 0: "namespace", 1: "type", 2: "class", 3: "enum", 4: "interface", 5: "struct", 6: "typeParameter", 7: "parameter", 8: "variable"
	// Decode tokens: they are stored as delta line, delta char, length, tokenType, tokenModifiers
	t.Logf("Tokens length: %d", len(tokens.Data))

	line := 0
	char := 0
	for i := 0; i < len(tokens.Data); i += 5 {
		deltaLine := tokens.Data[i]
		deltaChar := tokens.Data[i+1]
		length := tokens.Data[i+2]
		tokenType := tokens.Data[i+3]

		line += int(deltaLine)
		if deltaLine == 0 {
			char += int(deltaChar)
		} else {
			char = int(deltaChar)
		}

		// Find substring in content
		lines := strings.Split(content, "\n")
		var tokenStr string
		if line < len(lines) && char+int(length) <= len(lines[line]) {
			tokenStr = lines[line][char : char+int(length)]
		}

		t.Logf("Token at Line %d, Col %d (%s): Type %d (%s)", line, char, tokenStr, tokenType, legendTokenTypes[tokenType])
	}
}

func TestSemanticTokensGenericReceiver(t *testing.T) {
	handler := NewHandler()
	ctx := context.Background()

	uri := "file:///test.nr"
	content := `package main
type List[T] = struct {
    data: T[]
}
fn (self: #List[T]) Push[T](val: T) {
}`
	err := handler.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  uri,
			Text: content,
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	docVal, _ := handler.docs.Load(uri)
	doc := docVal.(*Document)
	t.Log("--- USES ---")
	for k, v := range doc.Info.Uses {
		t.Logf("Identifier %s (type %T, addr %p, line %d, col %d): %s (kind %v)", k.Value, k, k, k.Pos().Line, k.Pos().Column, v.Name, v.Kind)
	}
	t.Log("--- DEFS ---")
	for k, v := range doc.Info.Defs {
		t.Logf("Identifier %s (type %T, addr %p, line %d, col %d): %s (kind %v)", k.Value, k, k, k.Pos().Line, k.Pos().Column, v.Name, v.Kind)
	}

	tokens, err := handler.TextDocumentSemanticTokensFull(ctx, nil, &SemanticTokensParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("Semantic tokens failed: %v", err)
	}

	line := 0
	char := 0
	for i := 0; i < len(tokens.Data); i += 5 {
		deltaLine := tokens.Data[i]
		deltaChar := tokens.Data[i+1]
		length := tokens.Data[i+2]
		tokenType := tokens.Data[i+3]

		line += int(deltaLine)
		if deltaLine == 0 {
			char += int(deltaChar)
		} else {
			char = int(deltaChar)
		}

		lines := strings.Split(content, "\n")
		var tokenStr string
		if line < len(lines) && char+int(length) <= len(lines[line]) {
			tokenStr = lines[line][char : char+int(length)]
		}

		t.Logf("Token at Line %d, Col %d (%s): Type %d (%s)", line, char, tokenStr, tokenType, legendTokenTypes[tokenType])
	}
}

func TestLSPStdlibJson(t *testing.T) {
	handler := NewHandler()

	wd, _ := os.Getwd()
	root := filepath.Dir(filepath.Dir(wd))

	testFilePath := filepath.Join(root, "pkg", "cmd", "test", "stdlib_json_test", "stdlib_json_test.nr")
	contentBytes, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read stdlib_json_test.nr: %v", err)
	}

	uri := "file:///" + filepath.ToSlash(testFilePath)

	err = handler.TextDocumentDidOpen(context.Background(), nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  uri,
			Text: string(contentBytes),
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	d, ok := handler.docs.Load(uri)
	if !ok {
		t.Fatalf("Document not loaded")
	}
	doc := d.(*Document)
	if doc.Program == nil || doc.Info == nil {
		t.Fatalf("Analysis failed")
	}

	t.Logf("Found %d diagnostics", len(doc.Diags.Diagnostics))
	for _, diag := range doc.Diags.Diagnostics {
		t.Errorf("DIAGNOSTIC: line %d, col %d: %s", diag.Range.Start.Line, diag.Range.Start.Column, diag.Message)
	}
}

func TestSemanticTokensGenericFunctionName(t *testing.T) {
	handler := NewHandler()
	ctx := context.Background()

	uri := "file:///generic_fn.nr"
	content := `package collections

pub fn NewLinkedList[T]() @LinkedList[T] {
    return alloc LinkedList[T]{ head: none, tail: none, size: 0 }
}
`
	err := handler.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, Text: content},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	tokens, err := handler.TextDocumentSemanticTokensFull(ctx, nil, &SemanticTokensParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("Semantic tokens failed: %v", err)
	}

	line := 0
	char := 0
	found := false
	for i := 0; i < len(tokens.Data); i += 5 {
		deltaLine := tokens.Data[i]
		deltaChar := tokens.Data[i+1]
		length := tokens.Data[i+2]
		tokenType := tokens.Data[i+3]

		line += int(deltaLine)
		if deltaLine == 0 {
			char += int(deltaChar)
		} else {
			char = int(deltaChar)
		}

		lines := strings.Split(content, "\n")
		var tokenStr string
		if line < len(lines) && char+int(length) <= len(lines[line]) {
			tokenStr = lines[line][char : char+int(length)]
		}
		if tokenStr == "NewLinkedList" && tokenType == 11 { // function
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected semantic function token on NewLinkedList, tokens: %v", tokens.Data)
	}
}

func TestSemanticTokensInterpolatedString(t *testing.T) {
	handler := NewHandler()
	ctx := context.Background()

	uri := "file:///interpolation.nr"
	content := `package main

fn main() {
    var a = 1
    io.println("Result: ${a + 1} done")
}
`
	err := handler.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, Text: content},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	tokens, err := handler.TextDocumentSemanticTokensFull(ctx, nil, &SemanticTokensParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	})
	if err != nil {
		t.Fatalf("Semantic tokens failed: %v", err)
	}

	line := 0
	char := 0
	for i := 0; i < len(tokens.Data); i += 5 {
		deltaLine := tokens.Data[i]
		deltaChar := tokens.Data[i+1]
		length := tokens.Data[i+2]
		tokenType := tokens.Data[i+3]

		line += int(deltaLine)
		if deltaLine == 0 {
			char += int(deltaChar)
		} else {
			char = int(deltaChar)
		}

		lines := strings.Split(content, "\n")
		var tokenStr string
		if line < len(lines) && char+int(length) <= len(lines[line]) {
			tokenStr = lines[line][char : char+int(length)]
		}

		if tokenType == 17 { // string
			if strings.Contains(tokenStr, "${") || strings.Contains(tokenStr, "}") {
				t.Fatalf("string token must not cover interpolation delimiters, got %q", tokenStr)
			}
		}
		if tokenType == 19 { // operator — delimiters and expression punctuation
			continue
		}
		if tokenStr == "tota" {
			t.Fatalf("identifier token must cover full name \"total\", got partial %q", tokenStr)
		}
	}
}

func TestLSPGecsEntity(t *testing.T) {
	handler := NewHandler()
	ctx := context.Background()

	wd, _ := os.Getwd()
	root := filepath.Dir(filepath.Dir(wd))
	entityPath := filepath.Join(root, "examples", "port_gecs", "gecs", "src", "entity.nr")

	contentBytes, err := os.ReadFile(entityPath)
	if err != nil {
		t.Fatalf("Failed to read entity.nr: %v", err)
	}

	uri := "file:///" + filepath.ToSlash(entityPath)

	err = handler.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  uri,
			Text: string(contentBytes),
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	d, ok := handler.docs.Load(uri)
	if !ok {
		t.Fatalf("Document not loaded")
	}
	doc := d.(*Document)
	if doc.Program == nil || doc.Info == nil {
		t.Fatalf("Analysis failed")
	}

	t.Logf("Found %d diagnostics", len(doc.Diags.Diagnostics))
	for _, diag := range doc.Diags.Diagnostics {
		t.Errorf("DIAGNOSTIC: file %s, line %d, col %d: %s", diag.File, diag.Range.Start.Line, diag.Range.Start.Column, diag.Message)
	}
}

func TestLSPGecsATypes(t *testing.T) {
	handler := NewHandler()
	ctx := context.Background()

	wd, _ := os.Getwd()
	root := filepath.Dir(filepath.Dir(wd))
	aTypesPath := filepath.Join(root, "examples", "port_gecs", "gecs", "src", "a_types.nr")

	contentBytes, err := os.ReadFile(aTypesPath)
	if err != nil {
		t.Fatalf("Failed to read a_types.nr: %v", err)
	}

	uri := "file:///" + filepath.ToSlash(aTypesPath)

	err = handler.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:  uri,
			Text: string(contentBytes),
		},
	})
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	d, ok := handler.docs.Load(uri)
	if !ok {
		t.Fatalf("Document not loaded")
	}
	doc := d.(*Document)
	if doc.Program == nil || doc.Info == nil {
		t.Fatalf("Analysis failed")
	}

	t.Logf("Found %d diagnostics", len(doc.Diags.Diagnostics))
	for _, diag := range doc.Diags.Diagnostics {
		t.Errorf("DIAGNOSTIC: file %s, line %d, col %d: %s", diag.File, diag.Range.Start.Line, diag.Range.Start.Column, diag.Message)
	}
}
