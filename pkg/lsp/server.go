package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/jsonrpc2"
)

type Server struct {
	Handler *Handler
}

func NewServer() *Server {
	return &Server{
		Handler: NewHandler(),
	}
}

// stdioReadWriteCloser wraps os.Stdin and os.Stdout
type stdioReadWriteCloser struct {
	in  io.Reader
	out io.Writer
}

func (s *stdioReadWriteCloser) Read(p []byte) (int, error) {
	return s.in.Read(p)
}

func (s *stdioReadWriteCloser) Write(p []byte) (int, error) {
	return s.out.Write(p)
}

func (s *stdioReadWriteCloser) Close() error {
	return nil
}

func (s *Server) Run() error {
	stream := jsonrpc2.NewBufferedStream(&stdioReadWriteCloser{os.Stdin, os.Stdout}, jsonrpc2.VSCodeObjectCodec{})
	handler := jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(s.handleRequest))
	logger := log.New(os.Stderr, "[LSP] ", log.LstdFlags)
	conn := jsonrpc2.NewConn(context.Background(), stream, handler, jsonrpc2.LogMessages(logger))

	s.Handler.conn = conn

	<-conn.DisconnectNotify()
	return nil
}

func (s *Server) handleRequest(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (res interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			logPath := filepath.Join(os.TempDir(), "NORA_lsp_panic.log")
			if f, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); f != nil {
				f.WriteString(fmt.Sprintf("[%v] PANIC in handleRequest: %v\n", time.Now().Format(time.RFC3339), r))
				f.Close()
			}
			err = &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInternalError,
				Message: fmt.Sprintf("LSP internal panic: %v", r),
			}
		}
	}()

	switch req.Method {
	case "initialize":
		var params InitializeParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return s.Handler.Initialize(ctx, conn, &params)

	case "initialized":
		return nil, nil

	case "textDocument/didOpen":
		var params DidOpenTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return nil, s.Handler.TextDocumentDidOpen(ctx, conn, &params)

	case "textDocument/didChange":
		var params DidChangeTextDocumentParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return nil, s.Handler.TextDocumentDidChange(ctx, conn, &params)

	case "textDocument/didSave":
		return nil, nil

	case "textDocument/hover":
		var params HoverParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return s.Handler.TextDocumentHover(ctx, conn, &params)

	case "textDocument/definition":
		var params DefinitionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return s.Handler.TextDocumentDefinition(ctx, conn, &params)

	case "textDocument/completion":
		var params CompletionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return s.Handler.TextDocumentCompletion(ctx, conn, &params)

	case "textDocument/formatting":
		var params DocumentFormattingParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return s.Handler.TextDocumentFormatting(ctx, conn, &params)

	case "textDocument/semanticTokens/full":
		var params SemanticTokensParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return s.Handler.TextDocumentSemanticTokensFull(ctx, conn, &params)

	case "textDocument/signatureHelp":
		var params SignatureHelpParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return s.Handler.TextDocumentSignatureHelp(ctx, conn, &params)

	case "textDocument/documentSymbol":
		var params DocumentSymbolParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return s.Handler.TextDocumentDocumentSymbol(ctx, conn, &params)

	case "textDocument/references":
		var params ReferenceParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return s.Handler.TextDocumentReferences(ctx, conn, &params)

	case "textDocument/prepareRename":
		var params PrepareRenameParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return s.Handler.TextDocumentPrepareRename(ctx, conn, &params)

	case "textDocument/rename":
		var params RenameParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return s.Handler.TextDocumentRename(ctx, conn, &params)

	case "textDocument/codeAction":
		var params CodeActionParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return s.Handler.TextDocumentCodeAction(ctx, conn, &params)

	case "textDocument/inlayHint":
		var params InlayHintParams
		if err := json.Unmarshal(*req.Params, &params); err != nil {
			return nil, err
		}
		return s.Handler.TextDocumentInlayHint(ctx, conn, &params)

	case "shutdown":
		return nil, nil
	case "exit":
		os.Exit(0)
		return nil, nil
	}

	if strings.HasPrefix(req.Method, "$/") {
		return nil, nil
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: "Method not found"}
}
