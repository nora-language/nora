package lsp

import "fmt"
import "os"
import "context"
import "testing"

type mockConn struct {}
func (m *mockConn) Notify(ctx context.Context, method string, params interface{}) error { return nil }
func (m *mockConn) Call(ctx context.Context, method string, params interface{}, result interface{}) error { return nil }
func (m *mockConn) Close() error { return nil }

func TestLSPBug(t *testing.T) {
    h := NewHandler()
    ctx := context.Background()
    
    mainURI := "file:///e:/Project/Project%20Chronos/second/vscode-nora/example/src/main.nr"
    mainContent, _ := os.ReadFile("e:/Project/Project Chronos/second/vscode-nora/example/src/main.nr")
    
    h.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
        TextDocument: TextDocumentItem{
            URI: mainURI,
            Text: string(mainContent),
        },
    })

    // Simulate TextDocumentDidOpen for match.nr
    matchURI := "file:///e:/Project/Project%20Chronos/second/vscode-nora/example/src/match.nr"
    matchContent, _ := os.ReadFile("e:/Project/Project Chronos/second/vscode-nora/example/src/match.nr")
    
    h.TextDocumentDidOpen(ctx, nil, &DidOpenTextDocumentParams{
        TextDocument: TextDocumentItem{
            URI: matchURI,
            Text: string(matchContent),
        },
    })
    
    docInterface, _ := h.docs.Load(matchURI)
    if docInterface != nil {
        doc := docInterface.(*Document)
        for _, diag := range doc.Diags.Diagnostics {
            fmt.Println("DIAG:", diag.Message)
        }
    }
    
    // Check logs
    fmt.Println("Done analyzing match.nr")
}
