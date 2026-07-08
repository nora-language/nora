package lsp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHoverCreateSurface(t *testing.T) {
	h := NewHandler()
	ctx := context.Background()

	// Convert absolute path to file URI
	absPath := filepath.Join("..", "..", "nora_wgpu", "examples", "triangle", "main.nr")
	absPath, err := filepath.Abs(absPath)
	if err != nil {
		t.Fatalf("Could not get abs path: %v", err)
	}
	uri := "file:///" + strings.ReplaceAll(filepath.ToSlash(absPath), "C:/", "")
	// If it's on E drive, replace E:/ as well
	uri = "file:///" + strings.ReplaceAll(filepath.ToSlash(absPath), "E:/", "e:/")
    
	contentBytes, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("Could not read main.nr: %v", err)
	}
	content := string(contentBytes)

	h.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:     uri,
			Version: 1,
			Text:    content,
		},
	})

	// Hover over CreateSurface
    lineIndex := 0
    charIndex := 0
    lines := strings.Split(content, "\n")
    for i, line := range lines {
        if idx := strings.Index(line, "CreateSurface"); idx != -1 {
            lineIndex = i
            charIndex = idx + 2 // anywhere inside CreateSurface
            break
        }
    }

    t.Logf("Found CreateSurface at line %d, char %d", lineIndex, charIndex)

	res, err := h.TextDocumentHover(ctx, nil, &HoverParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     Position{Line: lineIndex, Character: charIndex},
	})
	if err != nil {
		t.Fatalf("Hover failed: %v", err)
	}
	if res == nil {
		t.Fatal("Expected hover result, got nil")
	}
	t.Logf("Hover Result:\n%s", res.Contents.Value)
}
