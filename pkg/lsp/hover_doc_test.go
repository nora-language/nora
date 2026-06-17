package lsp

import (
	"context"
	"testing"

	"github.com/nora-language/nora/pkg/parser/ast"
)

func TestHoverDoc(t *testing.T) {
	h := NewHandler()
	ctx := context.Background()

	uri := "file:///test.nr"
	content := `/// Point represents a 2D coordinate
type Point = struct {
    /// X coordinate
    x: i32
}

fn main() {
    var p = Point{x: 1}
    p.x
}`
	h.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:     uri,
			Version: 1,
			Text:    content,
		},
	})

	// Hover over Point in "var p = Point"
	// Point is at line 7 (0-indexed), char 12
	res, err := h.TextDocumentHover(ctx, nil, &HoverParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 7, Character: 12},
	})
	if err != nil {
		t.Fatalf("Hover Point failed: %v", err)
	}
	if res == nil {
		t.Fatal("Expected hover result for Point, got nil")
	}
	t.Logf("Hover Point Result: %s", res.Contents.Value)
	if !contains(res.Contents.Value, "Point represents a 2D coordinate") {
		t.Errorf("Hover missing doc comment for Point. Got: %s", res.Contents.Value)
	}

	// Hover over x in "p.x"
	// x is at line 8, char 6
	res, err = h.TextDocumentHover(ctx, nil, &HoverParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 8, Character: 6},
	})
	if err != nil {
		t.Fatalf("Hover x failed: %v", err)
	}
	if res == nil {
		t.Fatal("Expected hover result for x, got nil")
	}
	t.Logf("Hover x Result: %s", res.Contents.Value)
	if !contains(res.Contents.Value, "X coordinate") {
		t.Errorf("Hover missing doc comment for field x. Got: %s", res.Contents.Value)
	}
}

func TestHoverReceiverAndReturn(t *testing.T) {
	h := NewHandler()
	ctx := context.Background()

	uri := "file:///test.nr"
	content := `/// Point represents a 2D coordinate
type Point = struct {
    x: i32
}

fn (p: Point) GetX() Point {
    return p
}`
	h.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:     uri,
			Version: 1,
			Text:    content,
		},
	})

	// Hover over Point in receiver "(p: Point)"
	// Point is at line 5 (0-indexed), char 7
	res, err := h.TextDocumentHover(ctx, nil, &HoverParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 5, Character: 7},
	})
	if err != nil {
		t.Fatalf("Hover receiver Point failed: %v", err)
	}
	if res == nil {
		t.Fatal("Expected hover result for receiver Point, got nil")
	}
	if !contains(res.Contents.Value, "Point represents a 2D coordinate") {
		t.Errorf("Hover missing doc comment for receiver Point. Got: %s", res.Contents.Value)
	}

	// Hover over Point in return type "GetX() Point"
	// Point is at line 5, char 21
	res, err = h.TextDocumentHover(ctx, nil, &HoverParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 5, Character: 21},
	})
	if err != nil {
		t.Fatalf("Hover return type Point failed: %v", err)
	}
	if res == nil {
		t.Fatal("Expected hover result for return type Point, got nil")
	}
	if !contains(res.Contents.Value, "Point represents a 2D coordinate") {
		t.Errorf("Hover missing doc comment for return type Point. Got: %s", res.Contents.Value)
	}
}

func TestHoverEnumAndInterface(t *testing.T) {
	h := NewHandler()
	ctx := context.Background()

	uri := "file:///test.nr"
	content := `/// Status represents state
type Status = enum {
    Ready,
    Loading,
    Error(msg: str)
}

/// Speaker protocol
type Speaker = interface {
    fn SayHello() str
}

fn test(s: Status, sp: Speaker) {}`

	h.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:     uri,
			Version: 1,
			Text:    content,
		},
	})

	// Print AST comment groups to see their lines
	h.docs.Range(func(key, val interface{}) bool {
		doc := val.(*Document)
		if doc.Program != nil {
			for _, f := range doc.Program.Files {
				for _, c := range f.Comments {
					t.Logf("COMMENT GROUP: at line %d: %q", c.Pos().Line, c.Text())
				}
				for _, stmt := range f.Statements {
					if ts, ok := stmt.(*ast.TypeStatement); ok {
						t.Logf("STATEMENT: type %s at line %d, doc exists: %v", ts.Name.Value, ts.Pos().Line, ts.Doc != nil)
					}
				}
			}
		}
		return true
	})

	// Hover over Status in test signature "s: Status"
	// Line 12 (0-indexed), char 11
	res, err := h.TextDocumentHover(ctx, nil, &HoverParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 12, Character: 11},
	})
	if err != nil {
		t.Fatalf("Hover Status failed: %v", err)
	}
	if res == nil {
		t.Fatal("Expected hover result for Status, got nil")
	}
	t.Logf("Hover Status Result:\n%s", res.Contents.Value)
	expectedEnum := "type Status = enum {\n    Error(msg: str),\n    Loading,\n    Ready\n}"
	if !contains(res.Contents.Value, expectedEnum) {
		t.Errorf("Hover missing correct enum format. Expected to contain:\n%s\nGot:\n%s", expectedEnum, res.Contents.Value)
	}
	if !contains(res.Contents.Value, "Status represents state") {
		t.Errorf("Hover missing doc comment for Status")
	}

	// Hover over Speaker in test signature "sp: Speaker"
	// Line 12, char 23
	res, err = h.TextDocumentHover(ctx, nil, &HoverParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 12, Character: 23},
	})
	if err != nil {
		t.Fatalf("Hover Speaker failed: %v", err)
	}
	if res == nil {
		t.Fatal("Expected hover result for Speaker, got nil")
	}
	t.Logf("Hover Speaker Result:\n%s", res.Contents.Value)
	expectedInterface := "type Speaker = interface {\n    fn SayHello() str\n}"
	if !contains(res.Contents.Value, expectedInterface) {
		t.Errorf("Hover missing correct interface format. Expected to contain:\n%s\nGot:\n%s", expectedInterface, res.Contents.Value)
	}
	if !contains(res.Contents.Value, "Speaker protocol") {
		t.Errorf("Hover missing doc comment for Speaker")
	}
}

func TestHoverGenericStructLeaseFieldsKeepTypeArgs(t *testing.T) {
	h := NewHandler()
	ctx := context.Background()

	uri := "file:///listnode_hover.nr"
	content := `package collections

type ListNode[T] = struct {
    value: T
    prev: #ListNode[T]
    next: @ListNode[T]
}
`

	h.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, Text: content},
	})

	res, err := h.TextDocumentHover(ctx, nil, &HoverParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 2, Character: 10},
	})
	if err != nil {
		t.Fatalf("Hover ListNode failed: %v", err)
	}
	if res == nil {
		t.Fatal("Expected hover for ListNode, got nil")
	}
	t.Logf("Hover ListNode:\n%s", res.Contents.Value)

	if !contains(res.Contents.Value, "prev: #ListNode[T]") {
		t.Errorf("prev field should show #ListNode[T], got:\n%s", res.Contents.Value)
	}
	if !contains(res.Contents.Value, "next: @ListNode[T]") {
		t.Errorf("next field should show @ListNode[T], got:\n%s", res.Contents.Value)
	}
}

func TestHoverInterfaceDoesNotExpandReferencedTypes(t *testing.T) {
	h := NewHandler()
	ctx := context.Background()

	uri := "file:///iterator_hover.nr"
	content := `package collections

pub type Option[T] = enum {
    None,
    Some(val: T)
}

pub type Iterator[T] = interface {
    fn (self: Iterator[T]) Next() Option[T]
}

fn demo(it: Iterator[i32]) {}
`

	h.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{URI: uri, Text: content},
	})

	// Hover Iterator type definition
	res, err := h.TextDocumentHover(ctx, nil, &HoverParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 7, Character: 14},
	})
	if err != nil {
		t.Fatalf("Hover Iterator failed: %v", err)
	}
	if res == nil {
		t.Fatal("Expected hover for Iterator, got nil")
	}
	t.Logf("Hover Iterator:\n%s", res.Contents.Value)

	if !contains(res.Contents.Value, "type Iterator[T] = interface") {
		t.Errorf("Iterator hover should show interface definition, got:\n%s", res.Contents.Value)
	}
	if !contains(res.Contents.Value, "Next() Option[T]") {
		t.Errorf("Iterator hover should reference Option[T] by name, got:\n%s", res.Contents.Value)
	}
	if contains(res.Contents.Value, "type Option[T] = enum") {
		t.Errorf("Iterator hover must not inline-expand Option definition, got:\n%s", res.Contents.Value)
	}

	// Hover Option directly should still show full enum
	res, err = h.TextDocumentHover(ctx, nil, &HoverParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: 2, Character: 13},
	})
	if err != nil {
		t.Fatalf("Hover Option failed: %v", err)
	}
	if res == nil {
		t.Fatal("Expected hover for Option, got nil")
	}
	if !contains(res.Contents.Value, "type Option[T] = enum") {
		t.Errorf("Option hover should show full definition, got:\n%s", res.Contents.Value)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s[0:len(substr)] == substr || contains(s[1:], substr))
}
