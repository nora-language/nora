package lsp

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/DwiYI/Project-Nora/pkg/diag"
	"github.com/DwiYI/Project-Nora/pkg/format"
	"github.com/DwiYI/Project-Nora/pkg/lexer"
	"github.com/DwiYI/Project-Nora/pkg/parser"
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/plugin"
	"github.com/DwiYI/Project-Nora/pkg/semantic"
	"github.com/DwiYI/Project-Nora/pkg/token"
	"github.com/DwiYI/Project-Nora/pkg/topology"
	"github.com/DwiYI/Project-Nora/pkg/types"
	"github.com/sourcegraph/jsonrpc2"
	"gopkg.in/yaml.v3"
)

type Document struct {
	URI     string
	Content string
	Program *ast.Program
	Info    *semantic.SemanticInfo
	Diags   *diag.Collection
	mu      sync.RWMutex       // Protects Document fields
	sema    chan struct{}      // Analysis semaphore (prevents parallel analysis)
	cancel  context.CancelFunc // Current analysis cancellation
	timer   *time.Timer        // Debounce timer
}

type Handler struct {
	docs          sync.Map // URI -> *Document
	fileCache     sync.Map // Path -> *ast.File
	fileMeta      sync.Map // path -> os.FileInfo (for cache invalidation)
	instanceCache sync.Map // templatePointer_typeKey -> *ast.FunctionStatement
	packageCache  sync.Map // dirPath -> *packageCacheEntry
	conn          *jsonrpc2.Conn

	cacheMu      sync.Mutex
	pmCache      *plugin.PluginManager
	preludeCache *ast.File
}

type packageCacheEntry struct {
	Scope           *semantic.Scope
	CapturedSymbols map[string]*semantic.Symbol
	Files           []*ast.File
	ModTimes      map[string]time.Time
	DirModTime    time.Time
	MethodSymbols map[types.NRType]map[string]*semantic.Symbol
	FieldSymbols  map[*types.StructType]map[string]*semantic.Symbol
	FuncScopes    map[*ast.FunctionStatement]*semantic.Scope
	Scopes        map[ast.Node]*semantic.Scope
	Defs          map[*ast.Identifier]*semantic.Symbol
	Uses          map[*ast.Identifier]*semantic.Symbol
	Types         map[ast.Node]types.NRType
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Initialize(ctx context.Context, conn *jsonrpc2.Conn, params *InitializeParams) (*InitializeResult, error) {
	return &InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync:   1, // Full
			HoverProvider:      true,
			DefinitionProvider: true,
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{".", ":", "#", "@"},
			},
			SignatureHelpProvider: &SignatureHelpOptions{
				TriggerCharacters:   []string{"(", ","},
				RetriggerCharacters: []string{",", " "},
			},
			DocumentFormattingProvider: true,
			DocumentSymbolProvider:     true,
			ReferencesProvider:         true,
			RenameProvider: &RenameOptions{
				PrepareProvider: true,
			},
			CodeActionProvider: true,
			InlayHintProvider:  true,
			SemanticTokensProvider: &SemanticTokensOptions{
				Legend: SemanticTokensLegend{
					TokenTypes:     legendTokenTypes,
					TokenModifiers: legendTokenModifiers,
				},
				Full: true,
			},
		},
	}, nil
}

var legendTokenTypes = []string{
	"namespace", "type", "class", "enum", "interface", "struct", "typeParameter", "parameter", "variable", "property", "enumMember", "function", "method", "macro", "keyword", "modifier", "comment", "string", "number", "operator",
}

var legendTokenModifiers = []string{
	"declaration", "definition", "readonly", "static", "deprecated", "abstract", "async", "modification", "documentation", "defaultLibrary",
}

func (h *Handler) TextDocumentDidOpen(ctx context.Context, conn *jsonrpc2.Conn, params *DidOpenTextDocumentParams) error {
	doc := &Document{
		URI:     params.TextDocument.URI,
		Content: params.TextDocument.Text,
		sema:    make(chan struct{}, 1),
	}
	h.docs.Store(doc.URI, doc)
	h.analyze(ctx, conn, doc)
	return nil
}

func (h *Handler) TextDocumentDidChange(ctx context.Context, conn *jsonrpc2.Conn, params *DidChangeTextDocumentParams) error {
	if len(params.ContentChanges) == 0 {
		return nil
	}

	uri := params.TextDocument.URI
	if d, ok := h.docs.Load(uri); ok {
		doc := d.(*Document)
		doc.mu.Lock()
		doc.Content = params.ContentChanges[0].Text

		// 1. Cancel previous analysis
		if doc.cancel != nil {
			doc.cancel()
			doc.cancel = nil
		}

		// 2. Debounce new analysis
		if doc.timer != nil {
			doc.timer.Stop()
		}

		doc.timer = time.AfterFunc(100*time.Millisecond, func() {
			// Create a fresh context for the new analysis (cancellable by the next didChange)
			analysisCtx, cancel := context.WithCancel(context.Background())

			doc.mu.Lock()
			doc.cancel = cancel
			doc.mu.Unlock()

			h.analyze(analysisCtx, conn, doc)
		})
		doc.mu.Unlock()
	}
	return nil
}

func (h *Handler) analyze(ctx context.Context, conn *jsonrpc2.Conn, doc *Document) {
	logPath := filepath.Join(os.TempDir(), "NORA_lsp.log")

	defer func() {
		if r := recover(); r != nil {
			f, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if f != nil {
				fmt.Fprintf(f, "[%v] PANIC in analyze: %v\n%s\n", time.Now().Format(time.RFC3339), r, string(debug.Stack()))
				f.Close()
			}
		}
	}()

	// Wait for the semaphore OR cancellation
	select {
	case doc.sema <- struct{}{}:
		defer func() { <-doc.sema }()
	case <-ctx.Done():
		return
	}

	f, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%v] START analyze: %s\n", time.Now().Format(time.RFC3339), doc.URI)
		f.Close()
	}

	path := uriToPath(doc.URI)
	analyzer := semantic.NewAnalyzer()
	analyzer.Context = ctx // Pass context for cancellation
	l := lexer.New(doc.Content, path)
	l.Context = ctx
	l.Diagnostics = analyzer.Diagnostics
	p := parser.New(l)
	p.AllowNoPackage = false
	p.Context = ctx
	p.Diagnostics = analyzer.Diagnostics

	file := p.Parse(path)
	prog := &ast.Program{Files: []*ast.File{file}}

	// Derive a std/ root from the document URI (strip file:// and go up to find std/)
	// Fall back to current working directory
	stdDir := deriveStdDir(doc.URI)

	loader := &LSPFileLoader{
		Handler:      h,
		Analyzer:     analyzer,
		Program:      prog,
		StdDir:       stdDir,
		addedFiles:   make(map[string]bool),
		Dependencies: make(map[string]Dependency),
	}
	if projRoot := findProjectRoot(path); projRoot != "" {
		loader.loadManifest(projRoot)
	}

	h.cacheMu.Lock()
	if h.preludeCache == nil && !loader.NoCore {
		coreDir := filepath.Join(filepath.Dir(stdDir), "core")
		preludePath := filepath.Join(coreDir, "prelude.nr")
		if _, err := os.Stat(preludePath); err == nil {
			if contentBytes, err := os.ReadFile(preludePath); err == nil {
				preludeL := lexer.New(string(contentBytes), preludePath)
				preludeL.Context = ctx
				preludeL.Diagnostics = analyzer.Diagnostics
				preludeP := parser.New(preludeL)
				preludeP.AllowNoPackage = false
				preludeP.Context = ctx
				preludeP.Diagnostics = analyzer.Diagnostics
				h.preludeCache = preludeP.Parse(preludePath)
			}
		}
	}
	
	var preludeFile = h.preludeCache
	if preludeFile != nil {
		prog.Files = append(prog.Files, preludeFile)
	}

	if h.pmCache == nil {
		h.pmCache = plugin.NewPluginManager()
		h.pmCache.RegisterBuiltinMacros()
		
		// Load plugins from stdDir
		filepath.Walk(stdDir, func(p string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".wasm") {
				name := strings.TrimSuffix(info.Name(), ".wasm")
				h.pmCache.LoadPlugin(name, p)
			}
			return nil
		})
	}
	
	pm := h.pmCache
	
	// Load workspace plugins (these could change, so we do it dynamically, or they might error if already loaded, but pmCache ignores duplicate loads)
	if projRoot := findProjectRoot(path); projRoot != "" {
		for _, p := range loader.Plugins {
			name := filepath.Base(p)
			name = strings.TrimSuffix(name, filepath.Ext(name))
			pm.LoadPlugin(name, p)
		}
	}
	for _, file := range prog.Files {
		pm.ProcessMacroForFile(file)
	}
	h.cacheMu.Unlock()

	// Track the entry file
	if file.Name != "" {
		loader.addedFiles[file.Name] = true
	}
	if preludeFile != nil && preludeFile.Name != "" {
		loader.addedFiles[preludeFile.Name] = true
	}

	analyzer.Loader = loader

	// 1. Collect symbols from all files in the program (shallow pass)
	// This ensures we know about all types and functions across the package.
	analyzer.CollectSymbols(prog)

	// 1.5. Populate struct fields, enum variants, interface methods for all files
	for _, f := range prog.Files {
		analyzer.AnalyzeFileTypes(f)
	}

	// 2. Deep analyze ONLY the current file (expensive pass)
	// This performs type checking and body analysis.
	analyzer.Analyze(file)

	// 3. Run topology solver ONLY for the current file
	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.Solve(file)

	doc.Program = prog
	doc.Info = &analyzer.SemanticInfo
	doc.Diags = analyzer.Diagnostics

	f, _ = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "[%v] END analyze: %s (Diags: %d)\n", time.Now().Format(time.RFC3339), doc.URI, len(analyzer.Diagnostics.Diagnostics))
		f.Close()
	}

	h.publishDiagnostics(ctx, conn, doc)

	// Send refresh requests asynchronously to avoid blocking the handler
	go func() {
		conn.Call(ctx, "workspace/semanticTokens/refresh", nil, nil)
		conn.Call(ctx, "workspace/inlayHint/refresh", nil, nil)
	}()
}

func (h *Handler) publishDiagnostics(ctx context.Context, conn *jsonrpc2.Conn, doc *Document) {
	var diagnostics []Diagnostic

	logPath := filepath.Join(os.TempDir(), "NORA_lsp.log")
	f, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	docPath := uriToPath(doc.URI)

	doc.mu.RLock()
	var diags []diag.Diagnostic
	if doc.Diags != nil {
		diags = doc.Diags.Diagnostics
	}
	doc.mu.RUnlock()

	for _, d := range diags {
		if d.File != docPath {
			continue // Skip diagnostics from other files
		}

		severity := SeverityError
		switch d.Severity {
		case diag.Warning:
			severity = SeverityWarning
		case diag.Info:
			severity = SeverityInformation
		}

		if f != nil {
			fmt.Fprintf(f, "  DIAG [%s]: %s (line %d)\n", doc.URI, d.Message, d.Range.Start.Line)
		}

		diagnostics = append(diagnostics, Diagnostic{
			Range: Range{
				Start: Position{Line: d.Range.Start.Line - 1, Character: d.Range.Start.Column - 1},
				End:   Position{Line: d.Range.End.Line - 1, Character: d.Range.End.Column - 1},
			},
			Severity: severity,
			Source:   "Nora",
			Message:  d.Message,
		})
	}

	if f != nil {
		f.Close()
	}

	if diagnostics == nil {
		diagnostics = []Diagnostic{}
	}

	if conn != nil {
		conn.Notify(ctx, "textDocument/publishDiagnostics", PublishDiagnosticsParams{
			URI:         doc.URI,
			Diagnostics: diagnostics,
		})
	}
}

// ===================== HOVER =====================

// keywordDocs provides documentation for Nora keywords
var keywordDocs = map[string]string{
	"fn":       "Declares a function.\n\n```nora\nfn name(params) return_type { body }\n```",
	"var":      "Declares a variable with type inference.\n\n```nora\nvar name = value\nvar name: type = value\n```",
	"type":     "Defines a new named type (struct, sum type, or alias).\n\n```nora\ntype Name = struct { ... }\ntype Name = | Variant1 | Variant2\ntype Alias = i64\n```",
	"struct":   "Defines a struct type — a collection of named fields.\n\n```nora\ntype Point = struct {\n    X: f64,\n    Y: f64,\n}\n```",
	"protocol": "Defines a protocol (interface) — a contract that types can satisfy through structural typing. No `implements` keyword needed.\n\n```nora\nprotocol Speaker {\n    fn SayHello() str\n}\n```",
	"enum":     "Defines an enumeration type.",
	"spawn":    "Starts a new **Fiber** (lightweight cooperative task) on the M:N scheduler.\n\nThe spawned code runs concurrently. Shared data is automatically **Frozen** (read-only) by the compiler.\n\n```nora\nspawn { task() }\nspawn process(data)\n```",
	"parallel": "Runs child tasks on **separate OS thread workers** for true CPU parallelism.\n\nAll tasks inside the block execute physically in parallel and are automatically joined at the closing brace.\n\n```nora\nparallel {\n    task_a()\n    task_b()\n} // Automatic join\n```",
	"defer":    "Defers a function call until the current scope exits. Useful for resource cleanup.\n\n```nora\ndefer file.close()\n```",
	"pin":      "Pins a variable's **Lease**, preventing the compiler from freeing it. Required when passing data to C functions via FFI.\n\n```nora\npin buffer\nextern_call(buffer)\n```",
	"match":    "Pattern matching expression. All variants of a Sum Type must be handled (exhaustive).\n\n```nora\nmatch value {\n    Some(x) => { use(x) }\n    None    => { fallback() }\n}\n```",
	"select":   "Multiplexes on multiple channel operations. Similar to Go's `select`.\n\n```nora\nselect {\n    case msg = <-ch1:\n        handle(msg)\n    default:\n        fallback()\n}\n```",
	"chan":     "Channel type for communication between fibers. Data is transferred via **Lease Transfer** (zero-copy, only 8-byte permission pointer moves).",
	"alloc":    "Explicitly allocates a value on the **Lease-managed heap**. Normally the compiler decides placement automatically (stack vs heap).",
	"import":   "Imports a package. Circular dependencies are blocked at the Topological Analysis phase.\n\n```nora\nimport \"io\"\nimport net \"std/net\"\n```",
	"package":  "Declares the package name. All files in the same directory must use the same package name.\n\n```nora\npackage main\n```",
	"extern":   "Declares an external C function for FFI interop.\n\n```nora\nextern fn printf(fmt: ptr, ...) i32\n```",
	"export":   "Exports a Nora function so it can be called from C code (reverse FFI).",
	"for":      "Loop construct. Supports C-style, for-in, and infinite loops.\n\n```nora\nfor var i = 0; i < n; i++ { ... }\nfor item in collection { ... }\n```",
	"while":    "Loop that continues while a condition is true.\n\n```nora\nwhile condition { ... }\n```",
	"if":       "Conditional expression. Can be used as a statement or expression.\n\n```nora\nif condition {\n    ...\n} else {\n    ...\n}\n```",
	"return":   "Returns a value from the current function.",
	"break":    "Breaks out of the innermost loop.",
	"continue": "Skips to the next iteration of the innermost loop.",
	"pub":      "Makes a declaration visible outside the package.",
	"true":     "Boolean literal `true`.",
	"false":    "Boolean literal `false`.",
	"none":     "Represents the absence of a value (like null, but type-safe via `Option[T]`).",
}

func (h *Handler) TextDocumentHover(ctx context.Context, conn *jsonrpc2.Conn, params *HoverParams) (*Hover, error) {
	d, ok := h.docs.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	doc := d.(*Document)
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	if doc.Info == nil {
		return nil, nil
	}

	path := uriToPath(doc.URI)
	node := h.findNodeAt(doc.Program, path, params.Position)
	if node == nil {
		return nil, nil
	}

	// Priority 0: Keyword Hover — check if the word at cursor is a keyword
	if ident, ok := node.(*ast.Identifier); ok {
		if docText, found := keywordDocs[ident.Value]; found {
			// Only show keyword hover if it's NOT a symbol reference
			_, inUses := doc.Info.Uses[ident]
			_, inDefs := doc.Info.Defs[ident]
			if !inUses && !inDefs {
				return &Hover{
					Contents: MarkupContent{
						Kind:  "markdown",
						Value: fmt.Sprintf("**%s** (keyword)\n\n---\n\n%s", ident.Value, docText),
					},
				}, nil
			}
		}
	}

	// Priority 1: Identifiers (Symbol Hover provides Name, Kind, Type, and Doc)
	if ident, ok := node.(*ast.Identifier); ok {
		var sym *semantic.Symbol
		if s, ok := doc.Info.Uses[ident]; ok {
			sym = s
		} else if s, ok := doc.Info.Defs[ident]; ok {
			sym = s
		}

		if sym != nil {
			return &Hover{
				Contents: MarkupContent{
					Kind:  "markdown",
					Value: h.getSymbolHover(sym),
				},
			}, nil
		}
	}

	// Priority 1.5: Attribute Hover
	if attr, ok := node.(*ast.Attribute); ok {
		var docText string
		switch attr.Name {
		case "serialize":
			docText = "Generates JSON serialization and deserialization methods (`nr_serialize_json_...`, `nr_deserialize_json_...`) for this struct."
		case "inline":
			docText = "Instructs the compiler to inline this function at call sites."
		case "macro":
			docText = "Marks this function as a compiler built-in macro."
		case "builtin":
			docText = "Marks this function as a compiler built-in."
		case "shared":
			docText = "Marks this struct or protocol as shared across threads."
		default:
			docText = fmt.Sprintf("Attribute `%s`", attr.Name)
		}

		return &Hover{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("**[%s]** (attribute)\n\n---\n\n%s", attr.Name, docText),
			},
		}, nil
	}

	// Priority 2: Direct Type Hover (for literals, expressions, etc.)
	if t, ok := doc.Info.Types[node]; ok {
		return &Hover{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("```nora\n%s\n```", h.stringifyType(t)),
			},
		}, nil
	}

	return nil, nil
}

// ===================== DEFINITION =====================

func (h *Handler) TextDocumentDefinition(ctx context.Context, conn *jsonrpc2.Conn, params *DefinitionParams) (*Location, error) {
	d, ok := h.docs.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	doc := d.(*Document)
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	if doc.Info == nil {
		return nil, nil
	}

	path := uriToPath(doc.URI)
	node := h.findNodeAt(doc.Program, path, params.Position)
	if node == nil {
		return nil, nil
	}

	var defNode ast.Node
	if ident, ok := node.(*ast.Identifier); ok {
		if sym, ok := doc.Info.Uses[ident]; ok && sym != nil {
			defNode = sym.DefNode
		} else if sym, ok := doc.Info.Defs[ident]; ok && sym != nil {
			defNode = sym.DefNode
		}
	} else if _, ok := node.(*ast.Attribute); ok {
		// Attributes don't have standard definitions to jump to.
		return nil, nil
	}

	if defNode != nil {
		start := defNode.Pos()
		uri := pathToURI(start.Filename)
		if uri == "" {
			uri = doc.URI // Fallback to current doc if filename is empty (e.g. built-ins)
		}
		return &Location{
			URI: uri,
			Range: Range{
				Start: Position{Line: start.Line - 1, Character: start.Column - 1},
				End:   Position{Line: start.Line - 1, Character: start.Column - 1 + len(defNode.TokenLiteral())},
			},
		}, nil
	}

	return nil, nil
}

// ===================== COMPLETION =====================

func (h *Handler) TextDocumentCompletion(ctx context.Context, conn *jsonrpc2.Conn, params *CompletionParams) ([]CompletionItem, error) {
	d, ok := h.docs.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	doc := d.(*Document)
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	if doc.Info == nil {
		return nil, nil
	}

	// 1. Check for dot-access completion
	line := params.Position.Line
	char := params.Position.Character
	lines := strings.Split(doc.Content, "\n")
	if line < len(lines) {
		currentLine := lines[line]
		if char > 0 && char <= len(currentLine) {
			prefix := currentLine[:char]
			if strings.HasSuffix(prefix, ".") {
				// We are after a dot. Try to find the expression before the dot.
				path := uriToPath(doc.URI)
				nodeBeforeDot := h.findNodeAt(doc.Program, path, Position{Line: line, Character: char - 2})
				fmt.Fprintf(os.Stderr, "[DEBUG Completion] prefix: %q, path: %q, char: %d, nodeBeforeDot: %T (%+v)\n", prefix, path, char, nodeBeforeDot, nodeBeforeDot)
				if nodeBeforeDot != nil {
					fmt.Fprintf(os.Stderr, "[DEBUG Completion] nodeBeforeDot Pos: %+v\n", nodeBeforeDot.Pos())
					t, ok := doc.Info.Types[nodeBeforeDot]
					fmt.Fprintf(os.Stderr, "[DEBUG Completion] Types lookup: ok=%t, t=%v\n", ok, t)
					if ok && t != nil {
						return h.getMemberCompletions(t), nil
					}
				}
			}
		}
	}

	// 2. Standard scope-aware completion
	items := make(map[string]CompletionItem)

	scope := h.findScopeAt(doc, params.Position)
	for scope != nil {
		for name, sym := range scope.Symbols {
			if _, exists := items[name]; !exists {
				item := CompletionItem{
					Label:  name,
					Kind:   h.getCompletionKind(sym),
					Detail: h.typeNameRef(sym.Type, make(map[types.NRType]bool)),
				}
				// Add documentation from doc comments
				if sym.DefNode != nil {
					docText := h.getDocComment(sym.DefNode)
					if docText != "" {
						item.Documentation = &MarkupContent{
							Kind:  "markdown",
							Value: docText,
						}
					}
				}
				// Sort: local symbols first
				if scope.Kind == "function" || scope.Kind == "block" {
					item.SortText = "0_" + name
				} else if scope.Kind == "package" {
					item.SortText = "1_" + name
				} else {
					item.SortText = "2_" + name
				}
				items[name] = item
			}
		}
		scope = scope.Parent
	}

	// 3. Add keyword completions (only when not after a dot)
	keywordCompletions := h.getKeywordCompletions()
	for _, kw := range keywordCompletions {
		if _, exists := items[kw.Label]; !exists {
			items[kw.Label] = kw
		}
	}

	// 4. Convert map to slice
	var result []CompletionItem
	for _, item := range items {
		result = append(result, item)
	}

	if result == nil {
		result = []CompletionItem{}
	}

	return result, nil
}

func (h *Handler) getKeywordCompletions() []CompletionItem {
	type kwSnippet struct {
		label      string
		insertText string
		detail     string
	}
	snippets := []kwSnippet{
		{"fn", "fn ${1:name}(${2}) ${3} {\n\t$0\n}", "Function declaration"},
		{"if", "if ${1:condition} {\n\t$0\n}", "If statement"},
		{"for", "for ${1:init}; ${2:cond}; ${3:post} {\n\t$0\n}", "For loop"},
		{"forin", "for ${1:item} in ${2:collection} {\n\t$0\n}", "For-in loop"},
		{"while", "while ${1:condition} {\n\t$0\n}", "While loop"},
		{"match", "match ${1:value} {\n\t${2:Pattern} => {\n\t\t$0\n\t}\n}", "Match expression"},
		{"spawn", "spawn {\n\t$0\n}", "Spawn fiber"},
		{"parallel", "parallel {\n\t$0\n}", "Parallel block"},
		{"select", "select {\n\tcase ${1:val} = <-${2:ch}:\n\t\t$0\n\tdefault:\n}", "Select on channels"},
		{"type struct", "type ${1:Name} = struct {\n\t${2:Field}: ${3:type},\n}", "Struct type"},
		{"type sum", "type ${1:Name} =\n\t| ${2:Variant1}\n\t| ${3:Variant2}", "Sum type (enum)"},
		{"protocol", "protocol ${1:Name} {\n\tfn ${2:Method}(${3}) ${4}\n}", "Protocol (interface)"},
		{"defer", "defer ${1:call}(${2})", "Defer function call"},
		{"import", "import \"${1:package}\"", "Import package"},
	}

	var items []CompletionItem
	for _, s := range snippets {
		items = append(items, CompletionItem{
			Label:            s.label,
			Kind:             CompletionItemKindSnippet,
			Detail:           s.detail,
			InsertText:       s.insertText,
			InsertTextFormat: InsertTextFormatSnippet,
			SortText:         "3_" + s.label, // Keywords sort after symbols
		})
	}

	// Simple keyword completions (no snippet)
	simpleKeywords := []string{
		"return", "break", "continue", "var", "package",
		"extern", "export", "pub", "pin", "alloc", "chan",
		"true", "false", "none",
	}
	for _, kw := range simpleKeywords {
		items = append(items, CompletionItem{
			Label:    kw,
			Kind:     CompletionItemKindKeyword,
			Detail:   "keyword",
			SortText: "4_" + kw,
		})
	}

	// Type completions
	typeKeywords := []string{
		"i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64",
		"f32", "f64", "bool", "str", "ptr", "void", "fiber",
		"rune", "byte", "int",
	}
	for _, t := range typeKeywords {
		items = append(items, CompletionItem{
			Label:    t,
			Kind:     CompletionItemKindClass,
			Detail:   "type",
			SortText: "5_" + t,
		})
	}

	return items
}

// ===================== FORMATTING =====================

func (h *Handler) TextDocumentFormatting(ctx context.Context, conn *jsonrpc2.Conn, params *DocumentFormattingParams) ([]TextEdit, error) {
	d, ok := h.docs.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	doc := d.(*Document)
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	path := uriToPath(doc.URI)
	l := lexer.New(doc.Content, path)
	p := parser.New(l)
	p.AllowNoPackage = false
	p.PreserveParentheses = true
	p.DisableMacros = true
	file := p.Parse(path)

	if p.Diagnostics.HasErrors() {
		return nil, nil
	}

	config, _ := format.LoadConfig("nora-fmt.yaml")
	// Override with LSP options if provided
	if params.Options.TabSize > 0 {
		config.IndentSize = params.Options.TabSize
	}
	config.UseTabs = !params.Options.InsertSpaces

	formatter := format.New(config)
	formatted := formatter.Format(file)

	// Return a single edit for the entire document
	// Note: Line/Char are 0-based in LSP.
	// We need to find the total number of lines in the original content.
	lines := strings.Split(doc.Content, "\n")
	lastLine := len(lines) - 1
	lastChar := 0
	if lastLine >= 0 {
		lastChar = len(lines[lastLine])
	}

	return []TextEdit{
		{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: lastLine, Character: lastChar},
			},
			NewText: formatted,
		},
	}, nil
}

// ===================== SEMANTIC TOKENS =====================

func (h *Handler) TextDocumentSemanticTokensFull(ctx context.Context, conn *jsonrpc2.Conn, params *SemanticTokensParams) (*SemanticTokens, error) {
	d, ok := h.docs.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	doc := d.(*Document)
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	if doc.Info == nil || doc.Program == nil {
		return nil, nil
	}

	path := uriToPath(doc.URI)
	var file *ast.File
	for _, f := range doc.Program.Files {
		if normalizePath(f.Name) == path {
			file = f
			break
		}
	}
	if file == nil {
		return nil, nil
	}

	type rawToken struct {
		line           int
		char           int
		length         int
		tokenType      int
		tokenModifiers int
	}
	var rawTokens []rawToken

	addToken := func(line, char, length, tokenType, tokenModifiers int) {
		rawTokens = append(rawTokens, rawToken{
			line:           line,
			char:           char,
			length:         length,
			tokenType:      tokenType,
			tokenModifiers: tokenModifiers,
		})
	}

	// Helper to map SymbolKind to Token Type index
	getTokenType := func(sym *semantic.Symbol) int {
		switch sym.Kind {
		case semantic.SymVar:
			return 8 // variable
		case semantic.SymFunc:
			return 11 // function
		case semantic.SymType:
			if _, ok := sym.Type.(*types.GenericType); ok {
				return 6 // typeParameter
			}
			if _, ok := sym.Type.(*types.StructType); ok {
				return 5 // struct
			}
			if _, ok := sym.Type.(*types.ProtocolType); ok {
				return 4 // interface
			}
			if _, ok := sym.Type.(*types.SumType); ok {
				return 3 // enum
			}
			return 1 // type
		case semantic.SymModule:
			return 0 // namespace
		case semantic.SymParam:
			return 7 // parameter
		case semantic.SymPackage:
			return 0 // namespace
		case semantic.SymVariant:
			return 10 // enumMember
		}
		return 8 // default variable
	}

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil || ast.IsNil(n) {
			return false
		}

		// Skip nodes that were injected from another file (e.g., serialization_generated.nr)
		if n.Pos().Filename != file.Name {
			return true // Continue to next node, but skip this one
		}

		// Handle Identifiers
		if ident, ok := n.(*ast.Identifier); ok {
			if ident.Token.Type == token.STR {
				pos := ident.Pos()
				length := getStringLiteralLength(doc.Content, ident.Token.Position.Offset)
				if length == 0 {
					length = len(ident.Token.Literal)
				}
				addToken(pos.Line-1, pos.Column-1, length, 17, 0) // string
				return true
			}

			var sym *semantic.Symbol
			if s, ok := doc.Info.Uses[ident]; ok {
				sym = s
			} else if s, ok := doc.Info.Defs[ident]; ok {
				sym = s
			}

			if sym != nil {
				pos := ident.Pos()
				tokenType := getTokenType(sym)
				length := identifierTokenLength(doc.Content, ident)
				addToken(pos.Line-1, pos.Column-1, length, tokenType, 0)
			}
		}

		// Handle other literals (optional but nice)
		switch lit := n.(type) {
		case *ast.IntegerLiteral, *ast.FloatLiteral:
			pos := lit.Pos()
			addToken(pos.Line-1, pos.Column-1, len(lit.TokenLiteral()), 18, 0) // number
		case *ast.InterpolatedString:
			addInterpolatedStringTokens(doc.Content, lit, addToken)
		case *ast.StringLiteral:
			// Parts inside InterpolatedString have no source token; handled by addInterpolatedStringTokens.
			if lit.Token.Type == "" {
				return true
			}
			pos := lit.Pos()
			length := getStringLiteralLength(doc.Content, lit.Token.Position.Offset)
			if length == 0 {
				length = len(lit.TokenLiteral())
			}
			addToken(pos.Line-1, pos.Column-1, length, 17, 0) // string
		case *ast.RuneLiteral:
			pos := lit.Pos()
			length := getStringLiteralLength(doc.Content, lit.Token.Position.Offset)
			if length == 0 {
				length = len(lit.TokenLiteral())
			}
			addToken(pos.Line-1, pos.Column-1, length, 17, 0) // string
		}

		return true
	})

	// Sort the raw tokens in proper document order: line first, char second.
	// This prevents negative offsets in delta calculation when AST nodes are traversed out of physical layout order.
	sort.Slice(rawTokens, func(i, j int) bool {
		if rawTokens[i].line != rawTokens[j].line {
			return rawTokens[i].line < rawTokens[j].line
		}
		return rawTokens[i].char < rawTokens[j].char
	})

	var tokens []uint32
	lastLine := 0
	lastChar := 0
	for _, rt := range rawTokens {
		deltaLine := uint32(rt.line - lastLine)
		deltaChar := uint32(rt.char)
		if deltaLine == 0 {
			deltaChar = uint32(rt.char - lastChar)
		}
		tokens = append(tokens, deltaLine, deltaChar, uint32(rt.length), uint32(rt.tokenType), uint32(rt.tokenModifiers))
		lastLine = rt.line
		lastChar = rt.char
	}

	return &SemanticTokens{
		Data: tokens,
	}, nil
}

// ===================== SIGNATURE HELP =====================

func (h *Handler) TextDocumentSignatureHelp(ctx context.Context, conn *jsonrpc2.Conn, params *SignatureHelpParams) (*SignatureHelp, error) {
	d, ok := h.docs.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	doc := d.(*Document)
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	if doc.Info == nil {
		return nil, nil
	}

	// Find the call expression containing the cursor
	path := uriToPath(doc.URI)
	line := params.Position.Line
	char := params.Position.Character
	lines := strings.Split(doc.Content, "\n")
	if line >= len(lines) {
		return nil, nil
	}

	// Walk backwards from cursor to find the function name and count commas for active parameter
	currentLine := lines[line]
	if char > len(currentLine) {
		char = len(currentLine)
	}

	// Count unclosed parens and commas to determine context
	parenDepth := 0
	commaCount := 0
	funcEnd := -1

	for i := char - 1; i >= 0; i-- {
		ch := currentLine[i]
		if ch == ')' {
			parenDepth++
		} else if ch == '(' {
			if parenDepth == 0 {
				funcEnd = i
				break
			}
			parenDepth--
		} else if ch == ',' && parenDepth == 0 {
			commaCount++
		}
	}

	if funcEnd < 0 {
		return nil, nil
	}

	// Find the node at the function name position
	node := h.findNodeAt(doc.Program, path, Position{Line: line, Character: funcEnd - 1})
	if node == nil {
		return nil, nil
	}

	// Resolve the function type
	var fnType *types.FunctionType
	var fnName string

	if ident, ok := node.(*ast.Identifier); ok {
		var sym *semantic.Symbol
		if s, ok := doc.Info.Uses[ident]; ok {
			sym = s
		} else if s, ok := doc.Info.Defs[ident]; ok {
			sym = s
		}
		if sym != nil {
			if ft, ok := sym.Type.(*types.FunctionType); ok {
				fnType = ft
				fnName = sym.Name
			}
		}
	}

	// Also check for method calls (selector expressions)
	if fnType == nil {
		if sel, ok := node.(*ast.SelectorExpression); ok {
			if t, ok := doc.Info.Types[sel]; ok {
				if ft, ok := t.(*types.FunctionType); ok {
					fnType = ft
					if sel.Field != nil {
						fnName = sel.Field.Value
					}
				}
			}
		}
	}

	if fnType == nil {
		return nil, nil
	}

	// Build signature label
	var paramLabels []string
	var paramInfo []ParameterInformation

	labelBuilder := fmt.Sprintf("fn %s(", fnName)
	offset := len(labelBuilder)

	// Try to resolve parameter names from the AST FunctionStatement
	var paramNames []string
	if ident, ok := node.(*ast.Identifier); ok {
		var sym *semantic.Symbol
		if s, ok := doc.Info.Uses[ident]; ok {
			sym = s
		} else if s, ok := doc.Info.Defs[ident]; ok {
			sym = s
		}
		if sym != nil {
			if fnStmt, ok := sym.DefNode.(*ast.FunctionStatement); ok {
				for _, p := range fnStmt.Parameters {
					if p.Name != nil {
						paramNames = append(paramNames, p.Name.Value)
					}
				}
			}
		}
	}

	for i, p := range fnType.Params {
		pStr := h.stringifyType(p)
		if i < len(fnType.ParamLeases) {
			switch fnType.ParamLeases[i] {
			case types.LeaseWrite:
				pStr = "#" + pStr
			case types.LeaseMove:
				pStr = "@" + pStr
			}
		}
		paramName := fmt.Sprintf("p%d", i)
		if i < len(paramNames) && paramNames[i] != "" {
			paramName = paramNames[i]
		}
		paramStr := fmt.Sprintf("%s: %s", paramName, pStr)

		startOffset := offset
		endOffset := offset + len(paramStr)

		paramInfo = append(paramInfo, ParameterInformation{
			Label: [2]int{startOffset, endOffset},
		})

		paramLabels = append(paramLabels, paramStr)
		offset = endOffset + 2 // ", "
	}

	label := labelBuilder + strings.Join(paramLabels, ", ") + ")"
	if fnType.Return != nil && fnType.Return.Name() != "void" {
		label += " " + h.stringifyType(fnType.Return)
	}

	return &SignatureHelp{
		Signatures: []SignatureInformation{
			{
				Label:      label,
				Parameters: paramInfo,
			},
		},
		ActiveSignature: 0,
		ActiveParameter: commaCount,
	}, nil
}

// ===================== DOCUMENT SYMBOLS =====================

func safeName(name string) string {
	if name == "" {
		return "<unnamed>"
	}
	return name
}

func (h *Handler) TextDocumentDocumentSymbol(ctx context.Context, conn *jsonrpc2.Conn, params *DocumentSymbolParams) ([]DocumentSymbol, error) {
	d, ok := h.docs.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	doc := d.(*Document)
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	if doc.Program == nil {
		return nil, nil
	}

	path := uriToPath(doc.URI)
	var file *ast.File
	for _, f := range doc.Program.Files {
		if normalizePath(f.Name) == path {
			file = f
			break
		}
	}
	if file == nil {
		return nil, nil
	}

	var symbols []DocumentSymbol

	for _, stmt := range file.Statements {
		switch s := stmt.(type) {
		case *ast.PackageStatement:
			if s.Name != nil {
				symbols = append(symbols, DocumentSymbol{
					Name:           safeName(s.Name.Value),
					Kind:           SymbolKindPackage,
					Range:          makeFullRange(s, s.Name.Pos(), len(s.Name.Value)),
					SelectionRange: posToRange(s.Name.Pos(), len(s.Name.Value)),
				})
			}

		case *ast.ImportStatement:
			importName := s.PathValue()
			if importName == "" {
				importName = "<import>"
			}
			if s.Alias != nil {
				importName = s.Alias.Value + " \"" + s.PathValue() + "\""
			}
			symbols = append(symbols, DocumentSymbol{
				Name:           importName,
				Kind:           SymbolKindModule,
				Range:          makeFullRange(s, s.Path.Pos(), len(s.PathValue())),
				SelectionRange: posToRange(s.Path.Pos(), len(s.PathValue())),
			})

		case *ast.FunctionStatement:
			if s.Name == nil {
				continue
			}
			detail := ""
			if doc.Info != nil {
				if sym, ok := doc.Info.Defs[s.Name]; ok && sym != nil {
					detail = h.stringifyType(sym.Type)
				}
			}
			kind := SymbolKindFunction
			if s.Receiver != nil {
				kind = SymbolKindMethod
			}

			funcSym := DocumentSymbol{
				Name:           safeName(s.Name.Value),
				Detail:         detail,
				Kind:           kind,
				Range:          makeFullRange(s, s.Name.Pos(), len(s.Name.Value)),
				SelectionRange: posToRange(s.Name.Pos(), len(s.Name.Value)),
			}

			// Add parameters as children
			for _, param := range s.Parameters {
				if param.Name != nil {
					paramDetail := ""
					if param.Type != nil {
						paramDetail = param.Type.String()
					}
					funcSym.Children = append(funcSym.Children, DocumentSymbol{
						Name:           safeName(param.Name.Value),
						Detail:         paramDetail,
						Kind:           SymbolKindVariable,
						Range:          makeFullRange(param, param.Name.Pos(), len(param.Name.Value)),
						SelectionRange: posToRange(param.Name.Pos(), len(param.Name.Value)),
					})
				}
			}

			symbols = append(symbols, funcSym)

		case *ast.TypeStatement:
			if s.Name == nil {
				continue
			}
			kind := SymbolKindClass
			detail := "type"

			if _, ok := s.Value.(*ast.StructLiteral); ok {
				kind = SymbolKindStruct
				detail = "struct"
			} else if _, ok := s.Value.(*ast.InterfaceLiteral); ok {
				kind = SymbolKindInterface
				detail = "protocol"
			} else if _, ok := s.Value.(*ast.SumTypeLiteral); ok {
				kind = SymbolKindEnum
				detail = "sum type"
			}

			typeSym := DocumentSymbol{
				Name:           safeName(s.Name.Value),
				Detail:         detail,
				Kind:           kind,
				Range:          makeFullRange(s, s.Name.Pos(), len(s.Name.Value)),
				SelectionRange: posToRange(s.Name.Pos(), len(s.Name.Value)),
			}

			// Add fields/variants as children
			if structLit, ok := s.Value.(*ast.StructLiteral); ok {
				for _, field := range structLit.Fields {
					if field.Name != nil {
						fieldDetail := ""
						if field.Type != nil {
							fieldDetail = field.Type.String()
						}
						typeSym.Children = append(typeSym.Children, DocumentSymbol{
							Name:           safeName(field.Name.Value),
							Detail:         fieldDetail,
							Kind:           SymbolKindField,
							Range:          makeFullRange(field, field.Name.Pos(), len(field.Name.Value)),
							SelectionRange: posToRange(field.Name.Pos(), len(field.Name.Value)),
						})
					}
				}
			} else if sumLit, ok := s.Value.(*ast.SumTypeLiteral); ok {
				for _, variant := range sumLit.Variants {
					if variant.Name != nil {
						typeSym.Children = append(typeSym.Children, DocumentSymbol{
							Name:           safeName(variant.Name.Value),
							Detail:         "variant",
							Kind:           SymbolKindEnumMember,
							Range:          makeFullRange(variant, variant.Name.Pos(), len(variant.Name.Value)),
							SelectionRange: posToRange(variant.Name.Pos(), len(variant.Name.Value)),
						})
					}
				}
			}

			symbols = append(symbols, typeSym)

		case *ast.VarStatement:
			if s.Name == nil {
				continue
			}
			detail := ""
			if doc.Info != nil {
				if sym, ok := doc.Info.Defs[s.Name]; ok && sym != nil {
					detail = h.stringifyType(sym.Type)
				}
			}
			symbols = append(symbols, DocumentSymbol{
				Name:           safeName(s.Name.Value),
				Detail:         detail,
				Kind:           SymbolKindVariable,
				Range:          makeFullRange(s, s.Name.Pos(), len(s.Name.Value)),
				SelectionRange: posToRange(s.Name.Pos(), len(s.Name.Value)),
			})
		}
	}

	if symbols == nil {
		symbols = []DocumentSymbol{}
	}

	return symbols, nil
}

// ===================== REFERENCES =====================

func (h *Handler) TextDocumentReferences(ctx context.Context, conn *jsonrpc2.Conn, params *ReferenceParams) ([]Location, error) {
	d, ok := h.docs.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	doc := d.(*Document)
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	if doc.Info == nil {
		return nil, nil
	}

	path := uriToPath(doc.URI)
	node := h.findNodeAt(doc.Program, path, params.Position)
	if node == nil {
		return nil, nil
	}

	// Find the target symbol
	var targetSym *semantic.Symbol
	if ident, ok := node.(*ast.Identifier); ok {
		if s, ok := doc.Info.Uses[ident]; ok {
			targetSym = s
		} else if s, ok := doc.Info.Defs[ident]; ok {
			targetSym = s
		}
	}

	if targetSym == nil {
		return nil, nil
	}

	var locations []Location

	// Include declaration
	if params.Context.IncludeDeclaration && targetSym.DefNode != nil {
		defPos := targetSym.DefNode.Pos()
		defURI := pathToURI(defPos.Filename)
		if defURI == "" {
			defURI = doc.URI
		}
		locations = append(locations, Location{
			URI: defURI,
			Range: Range{
				Start: Position{Line: defPos.Line - 1, Character: defPos.Column - 1},
				End:   Position{Line: defPos.Line - 1, Character: defPos.Column - 1 + len(targetSym.Name)},
			},
		})
	}

	// Search through all definitions
	for ident, sym := range doc.Info.Defs {
		if sym == targetSym {
			pos := ident.Pos()
			uri := pathToURI(pos.Filename)
			if uri == "" {
				uri = doc.URI
			}
			locations = append(locations, Location{
				URI: uri,
				Range: Range{
					Start: Position{Line: pos.Line - 1, Character: pos.Column - 1},
					End:   Position{Line: pos.Line - 1, Character: pos.Column - 1 + len(ident.Value)},
				},
			})
		}
	}

	// Search through all uses
	for ident, sym := range doc.Info.Uses {
		if sym == targetSym {
			pos := ident.Pos()
			uri := pathToURI(pos.Filename)
			if uri == "" {
				uri = doc.URI
			}
			locations = append(locations, Location{
				URI: uri,
				Range: Range{
					Start: Position{Line: pos.Line - 1, Character: pos.Column - 1},
					End:   Position{Line: pos.Line - 1, Character: pos.Column - 1 + len(ident.Value)},
				},
			})
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	var unique []Location
	for _, loc := range locations {
		key := fmt.Sprintf("%s:%d:%d", loc.URI, loc.Range.Start.Line, loc.Range.Start.Character)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, loc)
		}
	}

	if unique == nil {
		unique = []Location{}
	}

	return unique, nil
}

// ===================== RENAME =====================

func (h *Handler) TextDocumentPrepareRename(ctx context.Context, conn *jsonrpc2.Conn, params *PrepareRenameParams) (*PrepareRenameResult, error) {
	d, ok := h.docs.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	doc := d.(*Document)
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	if doc.Info == nil {
		return nil, nil
	}

	path := uriToPath(doc.URI)
	node := h.findNodeAt(doc.Program, path, params.Position)
	if node == nil {
		return nil, nil
	}

	if ident, ok := node.(*ast.Identifier); ok {
		var sym *semantic.Symbol
		if s, ok := doc.Info.Uses[ident]; ok {
			sym = s
		} else if s, ok := doc.Info.Defs[ident]; ok {
			sym = s
		}
		if sym != nil {
			pos := ident.Pos()
			return &PrepareRenameResult{
				Range: Range{
					Start: Position{Line: pos.Line - 1, Character: pos.Column - 1},
					End:   Position{Line: pos.Line - 1, Character: pos.Column - 1 + len(ident.Value)},
				},
				Placeholder: ident.Value,
			}, nil
		}
	}

	return nil, nil
}

func (h *Handler) TextDocumentRename(ctx context.Context, conn *jsonrpc2.Conn, params *RenameParams) (*WorkspaceEdit, error) {
	d, ok := h.docs.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	doc := d.(*Document)
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	if doc.Info == nil {
		return nil, nil
	}

	path := uriToPath(doc.URI)
	node := h.findNodeAt(doc.Program, path, params.Position)
	if node == nil {
		return nil, nil
	}

	// Find the target symbol
	var targetSym *semantic.Symbol
	if ident, ok := node.(*ast.Identifier); ok {
		if s, ok := doc.Info.Uses[ident]; ok {
			targetSym = s
		} else if s, ok := doc.Info.Defs[ident]; ok {
			targetSym = s
		}
	}

	if targetSym == nil {
		return nil, nil
	}

	changes := make(map[string][]TextEdit)

	// Collect all occurrences (defs + uses)
	for ident, sym := range doc.Info.Defs {
		if sym == targetSym {
			pos := ident.Pos()
			uri := pathToURI(pos.Filename)
			if uri == "" {
				uri = doc.URI
			}
			changes[uri] = append(changes[uri], TextEdit{
				Range: Range{
					Start: Position{Line: pos.Line - 1, Character: pos.Column - 1},
					End:   Position{Line: pos.Line - 1, Character: pos.Column - 1 + len(ident.Value)},
				},
				NewText: params.NewName,
			})
		}
	}

	for ident, sym := range doc.Info.Uses {
		if sym == targetSym {
			pos := ident.Pos()
			uri := pathToURI(pos.Filename)
			if uri == "" {
				uri = doc.URI
			}
			changes[uri] = append(changes[uri], TextEdit{
				Range: Range{
					Start: Position{Line: pos.Line - 1, Character: pos.Column - 1},
					End:   Position{Line: pos.Line - 1, Character: pos.Column - 1 + len(ident.Value)},
				},
				NewText: params.NewName,
			})
		}
	}

	return &WorkspaceEdit{Changes: changes}, nil
}

// ===================== CODE ACTIONS =====================

func (h *Handler) TextDocumentCodeAction(ctx context.Context, conn *jsonrpc2.Conn, params *CodeActionParams) ([]CodeAction, error) {
	d, ok := h.docs.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	doc := d.(*Document)
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	if doc.Diags == nil {
		return nil, nil
	}

	var actions []CodeAction

	for _, diag := range params.Context.Diagnostics {
		// Quick fix: Add missing import
		if strings.Contains(diag.Message, "could not load package") {
			// Extract package name from error message
			parts := strings.Split(diag.Message, "\"")
			if len(parts) >= 2 {
				pkgName := parts[1]
				actions = append(actions, CodeAction{
					Title:       fmt.Sprintf("Add import \"%s\"", pkgName),
					Kind:        "quickfix",
					Diagnostics: []Diagnostic{diag},
					Edit: &WorkspaceEdit{
						Changes: map[string][]TextEdit{
							doc.URI: {
								{
									Range: Range{
										Start: Position{Line: 1, Character: 0},
										End:   Position{Line: 1, Character: 0},
									},
									NewText: fmt.Sprintf("import \"%s\"\n", pkgName),
								},
							},
						},
					},
				})
			}
		}

		// Quick fix: Remove unused variable
		if strings.Contains(diag.Message, "declared and not used") || strings.Contains(diag.Message, "unused") {
			actions = append(actions, CodeAction{
				Title:       "Remove unused variable",
				Kind:        "quickfix",
				Diagnostics: []Diagnostic{diag},
			})
		}
	}

	if actions == nil {
		actions = []CodeAction{}
	}

	return actions, nil
}

// ===================== INLAY HINTS =====================

func (h *Handler) TextDocumentInlayHint(ctx context.Context, conn *jsonrpc2.Conn, params *InlayHintParams) ([]InlayHint, error) {
	d, ok := h.docs.Load(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	doc := d.(*Document)
	doc.mu.RLock()
	defer doc.mu.RUnlock()
	if doc.Info == nil || doc.Program == nil {
		return nil, nil
	}

	path := uriToPath(doc.URI)
	var file *ast.File
	for _, f := range doc.Program.Files {
		if normalizePath(f.Name) == path {
			file = f
			break
		}
	}
	if file == nil {
		return nil, nil
	}

	var hints []InlayHint

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil || ast.IsNil(n) {
			return false
		}

		// Skip nodes that were injected from another file (e.g., serialization_generated.nr)
		if n.Pos().Filename != file.Name {
			return true // Continue to next node, but skip this one
		}

		// Inlay hint for inferred variable types: var x = 10 → show ": i32"
		if varStmt, ok := n.(*ast.VarStatement); ok {
			// Only show if no explicit type annotation
			if varStmt.Type == nil && varStmt.Name != nil {
				if sym, ok := doc.Info.Defs[varStmt.Name]; ok && sym != nil && sym.Type != nil {
					typeName := h.stringifyType(sym.Type)
					typeName = strings.ReplaceAll(typeName, "\r", " ")
					typeName = strings.ReplaceAll(typeName, "\n", " ")
					for strings.Contains(typeName, "  ") {
						typeName = strings.ReplaceAll(typeName, "  ", " ")
					}
					if typeName != "" && typeName != "unknown" && typeName != "void" {
						namePos := varStmt.Name.Pos()
						// Only show hints within the requested range
						hintLine := namePos.Line - 1
						if hintLine >= params.Range.Start.Line && hintLine <= params.Range.End.Line {
							hints = append(hints, InlayHint{
								Position: Position{
									Line:      hintLine,
									Character: namePos.Column - 1 + len(varStmt.Name.Value),
								},
								Label:        ": " + typeName,
								Kind:         InlayHintKindType,
								PaddingLeft:  false,
								PaddingRight: true,
							})
						}
					}
				}
			}
		}

		return true
	})

	if hints == nil {
		hints = []InlayHint{}
	}

	return hints, nil
}

// ===================== HELPERS =====================

func (h *Handler) getMemberCompletions(t types.NRType) []CompletionItem {
	var items []CompletionItem

	// Auto-dereference pointers
	if pt, ok := t.(*types.PointerType); ok {
		t = pt.Base
	}

	switch ct := t.(type) {
	case *semantic.ModuleType:
		for name, sym := range ct.Exports.Symbols {
			if sym.Visible != semantic.Public {
				continue
			}
			item := CompletionItem{
				Label:  name,
				Kind:   h.getCompletionKind(sym),
				Detail: h.typeNameRef(sym.Type, make(map[types.NRType]bool)),
			}
			if sym.DefNode != nil {
				docText := h.getDocComment(sym.DefNode)
				if docText != "" {
					item.Documentation = &MarkupContent{
						Kind:  "markdown",
						Value: docText,
					}
				}
			}
			items = append(items, item)
		}
	case *types.StructType:
		// Fields
		for name, ft := range ct.Fields {
			items = append(items, CompletionItem{
				Label:  name,
				Kind:   CompletionItemKindProperty,
				Detail: ft.Name(),
			})
		}
		// Methods
		for name, mt := range ct.Methods {
			items = append(items, CompletionItem{
				Label:  name,
				Kind:   CompletionItemKindMethod,
				Detail: mt.Name(),
			})
		}
	case *types.ProtocolType:
		for name, mt := range ct.Methods {
			items = append(items, CompletionItem{
				Label:  name,
				Kind:   CompletionItemKindMethod,
				Detail: mt.Name(),
			})
		}
	case *types.SumType:
		for name, variant := range ct.Variants {
			items = append(items, CompletionItem{
				Label:  name,
				Kind:   CompletionItemKindEnumMember,
				Detail: variant.Name,
			})
		}
		// Sum type methods
		if ct.Methods != nil {
			for name, mt := range ct.Methods {
				items = append(items, CompletionItem{
					Label:  name,
					Kind:   CompletionItemKindMethod,
					Detail: mt.Name(),
				})
			}
		}
	}

	return items
}

func (h *Handler) getCompletionKind(sym *semantic.Symbol) CompletionItemKind {
	switch sym.Kind {
	case semantic.SymFunc:
		return CompletionItemKindFunction
	case semantic.SymVar, semantic.SymParam:
		return CompletionItemKindVariable
	case semantic.SymType:
		if _, ok := sym.Type.(*types.StructType); ok {
			return CompletionItemKindStruct
		}
		if _, ok := sym.Type.(*types.ProtocolType); ok {
			return CompletionItemKindInterface
		}
		if _, ok := sym.Type.(*types.SumType); ok {
			return CompletionItemKindEnum
		}
		return CompletionItemKindClass
	case semantic.SymPackage:
		return CompletionItemKindModule
	case semantic.SymVariant:
		return CompletionItemKindEnumMember
	default:
		return CompletionItemKindText
	}
}

func (h *Handler) getSymbolHover(sym *semantic.Symbol) string {
	kindStr := sym.Kind.String()
	typeStr := h.stringifyType(sym.Type)

	res := fmt.Sprintf("**%s** (%s)\n\n", sym.Name, kindStr)
	res += fmt.Sprintf("```nora\n%s\n```", typeStr)

	// Show lease information
	if sym.LeaseKind == types.LeaseWrite {
		res += "\n\n🔒 **Lease**: Mutable borrow (`#`) — exclusive write access"
	} else if sym.LeaseKind == types.LeaseMove {
		res += "\n\n📦 **Lease**: Move (`@`) — ownership transferred"
	}

	// Show source package/file
	if sym.DefNode != nil && sym.DefNode.Pos().Filename != "" {
		filename := filepath.Base(sym.DefNode.Pos().Filename)
		res += fmt.Sprintf("\n\n📂 *Defined in `%s`*", filename)
	}

	// Add Doc Comments if available
	if sym.DefNode != nil {
		doc := h.getDocComment(sym.DefNode)
		if doc != "" {
			res += "\n\n---\n" + doc
		}
	}

	return res
}

func (h *Handler) getDocComment(node ast.Node) string {
	if node == nil || ast.IsNil(node) {
		return ""
	}

	var docGroup *ast.CommentGroup
	switch s := node.(type) {
	case *ast.FunctionStatement:
		docGroup = s.Doc
	case *ast.TypeStatement:
		docGroup = s.Doc
	case *ast.VarStatement:
		docGroup = s.Doc
	case *ast.PackageStatement:
		docGroup = s.Doc
	case *ast.FieldDefinition:
		docGroup = s.Doc
	case *ast.VariantDefinition:
		docGroup = s.Doc
	}

	if docGroup != nil {
		return docGroup.Text()
	}
	return ""
}

func (h *Handler) stringifyType(t types.NRType) string {
	return h.stringifyTypeRecursive(t, make(map[types.NRType]bool), true)
}

// typeNameRef renders a type by name (e.g. Option[T]) without expanding its full definition.
func (h *Handler) typeNameRef(t types.NRType, visited map[types.NRType]bool) string {
	if t == nil {
		return "unknown"
	}
	switch ft := t.(type) {
	case *types.StructType:
		if visited[ft] {
			return h.formatNamedType(ft.TypeName, ft.TypeParams, ft.TypeArgs, visited)
		}
		visited[ft] = true
		defer delete(visited, ft)
		return h.formatNamedType(ft.TypeName, ft.TypeParams, ft.TypeArgs, visited)
	case *types.SumType:
		if visited[ft] {
			return h.formatNamedType(ft.TypeName, ft.TypeParams, ft.TypeArgs, visited)
		}
		visited[ft] = true
		defer delete(visited, ft)
		return h.formatNamedType(ft.TypeName, ft.TypeParams, ft.TypeArgs, visited)
	case *types.ProtocolType:
		if visited[ft] {
			return h.formatNamedType(ft.ProtocolName, ft.TypeParams, ft.TypeArgs, visited)
		}
		visited[ft] = true
		defer delete(visited, ft)
		return h.formatNamedType(ft.ProtocolName, ft.TypeParams, ft.TypeArgs, visited)
	case *types.FunctionType:
		return h.stringifyFunctionType(ft, visited)
	case *types.GenericType:
		return ft.TypeParam
	case *types.PointerType:
		return h.formatPointerTypeRef(ft, visited)
	case *types.ListType:
		return "List[" + h.typeNameRef(ft.ElementType, visited) + "]"
	case *types.MapType:
		return "Map[" + h.typeNameRef(ft.Key, visited) + ", " + h.typeNameRef(ft.Value, visited) + "]"
	case *types.ChanType:
		return "chan " + h.typeNameRef(ft.Elem, visited)
	default:
		return t.Name()
	}
}

func (h *Handler) formatPointerTypeRef(pt *types.PointerType, visited map[types.NRType]bool) string {
	prefix := ""
	if pt.Leased {
		switch pt.Kind {
		case types.LeaseRead:
			prefix = "#"
		case types.LeaseWrite:
			prefix = "&"
		case types.LeaseMove:
			prefix = "@"
		}
	}
	base := h.typeNameRef(pt.Base, visited)
	if pt.IsArray {
		return prefix + "(" + base + ")[]"
	}
	return prefix + base
}

func (h *Handler) formatNamedType(name string, typeParams []*types.TypeParam, typeArgs []types.NRType, visited map[types.NRType]bool) string {
	if len(typeArgs) > 0 {
		args := make([]string, len(typeArgs))
		for i, a := range typeArgs {
			args[i] = h.typeNameRef(a, visited)
		}
		return name + "[" + strings.Join(args, ", ") + "]"
	}
	if len(typeParams) > 0 {
		params := make([]string, len(typeParams))
		for i, tp := range typeParams {
			params[i] = tp.Name
		}
		return name + "[" + strings.Join(params, ", ") + "]"
	}
	return name
}

func (h *Handler) stringifyFunctionType(ft *types.FunctionType, visited map[types.NRType]bool) string {
	var params []string
	for i, p := range ft.Params {
		pStr := h.typeNameRef(p, visited)
		if i < len(ft.ParamLeases) {
			switch ft.ParamLeases[i] {
			case types.LeaseWrite:
				pStr = "#" + pStr
			case types.LeaseMove:
				pStr = "@" + pStr
			}
		}
		params = append(params, pStr)
	}
	res := fmt.Sprintf("fn(%s)", strings.Join(params, ", "))
	if ft.Return != nil && ft.Return.Name() != "void" {
		res += " " + h.typeNameRef(ft.Return, visited)
	}
	return res
}

func (h *Handler) formatInterfaceMethod(name string, ft *types.FunctionType, owner *types.ProtocolType, visited map[types.NRType]bool) string {
	var b strings.Builder
	b.WriteString("fn")
	if ft.Receiver != nil {
		recvStr := h.typeNameRef(ft.Receiver, visited)
		if owner != nil {
			if recvPT, ok := ft.Receiver.(*types.ProtocolType); ok && recvPT.ProtocolName == owner.ProtocolName {
				recvStr = h.formatNamedType(owner.ProtocolName, owner.TypeParams, owner.TypeArgs, visited)
			}
		}
		b.WriteString(" (self: ")
		b.WriteString(recvStr)
		b.WriteString(")")
	}
	b.WriteString(" ")
	b.WriteString(name)
	if len(ft.Params) > 0 {
		var params []string
		for i, p := range ft.Params {
			pStr := h.typeNameRef(p, visited)
			if i < len(ft.ParamLeases) {
				switch ft.ParamLeases[i] {
				case types.LeaseWrite:
					pStr = "#" + pStr
				case types.LeaseMove:
					pStr = "@" + pStr
				}
			}
			params = append(params, pStr)
		}
		b.WriteString("(")
		b.WriteString(strings.Join(params, ", "))
		b.WriteString(")")
	} else {
		b.WriteString("()")
	}
	if ft.Return != nil && ft.Return.Name() != "void" {
		b.WriteString(" ")
		b.WriteString(h.typeNameRef(ft.Return, visited))
	}
	return b.String()
}

func (h *Handler) stringifyTypeRecursive(t types.NRType, visited map[types.NRType]bool, expand bool) string {
	if t == nil {
		return "unknown"
	}
	if !expand {
		return h.typeNameRef(t, visited)
	}
	if visited[t] {
		return h.typeNameRef(t, visited)
	}
	visited[t] = true
	defer delete(visited, t)

	switch ft := t.(type) {
	case *types.FunctionType:
		return h.stringifyFunctionType(ft, visited)

	case *types.StructType:
		typeHeader := h.formatNamedType(ft.TypeName, ft.TypeParams, ft.TypeArgs, visited)

		if len(ft.Fields) == 0 {
			return fmt.Sprintf("type %s = struct {}", typeHeader)
		}

		var fieldNames []string
		for name := range ft.Fields {
			fieldNames = append(fieldNames, name)
		}
		sort.Strings(fieldNames)

		var fields []string
		for _, name := range fieldNames {
			fieldType := ft.Fields[name]
			fields = append(fields, fmt.Sprintf("    %s: %s", name, h.typeNameRef(fieldType, visited)))
		}
		return fmt.Sprintf("type %s = struct {\n%s\n}", typeHeader, strings.Join(fields, "\n"))

	case *types.SumType:
		typeHeader := h.formatNamedType(ft.TypeName, ft.TypeParams, ft.TypeArgs, visited)

		if len(ft.Variants) == 0 {
			return fmt.Sprintf("type %s = enum {}", typeHeader)
		}

		var variantNames []string
		for name := range ft.Variants {
			variantNames = append(variantNames, name)
		}
		sort.Strings(variantNames)

		var variants []string
		for _, name := range variantNames {
			variant := ft.Variants[name]
			if len(variant.Fields) > 0 {
				var vFields []string
				if len(variant.FieldNames) > 0 {
					for _, fn := range variant.FieldNames {
						if fv, ok := variant.Fields[fn]; ok {
							vFields = append(vFields, fmt.Sprintf("%s: %s", fn, h.typeNameRef(fv, visited)))
						}
					}
				} else {
					var fnList []string
					for fn := range variant.Fields {
						fnList = append(fnList, fn)
					}
					sort.Strings(fnList)
					for _, fn := range fnList {
						fv := variant.Fields[fn]
						vFields = append(vFields, fmt.Sprintf("%s: %s", fn, h.typeNameRef(fv, visited)))
					}
				}
				variants = append(variants, fmt.Sprintf("    %s(%s)", name, strings.Join(vFields, ", ")))
			} else {
				variants = append(variants, fmt.Sprintf("    %s", name))
			}
		}
		return fmt.Sprintf("type %s = enum {\n%s\n}", typeHeader, strings.Join(variants, ",\n"))

	case *types.ProtocolType:
		typeHeader := h.formatNamedType(ft.ProtocolName, ft.TypeParams, ft.TypeArgs, visited)

		if len(ft.Methods) == 0 {
			return fmt.Sprintf("type %s = interface {}", typeHeader)
		}

		var methodNames []string
		for name := range ft.Methods {
			methodNames = append(methodNames, name)
		}
		sort.Strings(methodNames)

		var methods []string
		for _, name := range methodNames {
			mt := ft.Methods[name]
			methods = append(methods, "    "+h.formatInterfaceMethod(name, mt, ft, visited))
		}
		return fmt.Sprintf("type %s = interface {\n%s\n}", typeHeader, strings.Join(methods, "\n"))

	default:
		return t.Name()
	}
}

func (h *Handler) findNodeAt(prog *ast.Program, filename string, pos Position) ast.Node {
	line := pos.Line + 1
	col := pos.Character + 1
	normFilename := normalizePath(filename)

	var found ast.Node
	ast.Inspect(prog, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		// Handle typed nils
		if ast.IsNil(n) {
			return false
		}

		start := n.Pos()
		// Only consider nodes in the same file
		if normalizePath(start.Filename) != normFilename {
			return true // Continue to children (though they should be in the same file)
		}

		// If the node starts AFTER the cursor, skip it
		if start.Line > line || (start.Line == line && start.Column > col) {
			return false
		}

		// Calculate an approximate end position.
		// For leaf nodes, we can use the token literal length.
		// For composite nodes, we don't have an easy way without End(),
		// so we'll assume they contain the position if they start before it.
		// We'll prioritize the node that starts LATEST.

		isMatch := false
		if start.Line == line && col >= start.Column {
			// Check if it's a leaf node and if it ends before the cursor.
			// We only do this for nodes that don't have children in Inspect.
			isLeaf := false
			switch n.(type) {
			case *ast.Identifier, *ast.IntegerLiteral, *ast.StringLiteral, *ast.Boolean, *ast.FloatLiteral, *ast.ImaginaryLiteral, *ast.NoneLiteral:
				isLeaf = true
			}

			if isLeaf {
				length := len(n.TokenLiteral())
				if length > 0 && col > start.Column+length {
					return false
				}
			}
			isMatch = true
		} else if start.Line < line {
			// It starts on a previous line, could be a multi-line node
			isMatch = true
		}

		if isMatch {
			if found == nil {
				found = n
			} else {
				fStart := found.Pos()
				if start.Line > fStart.Line || (start.Line == fStart.Line && start.Column >= fStart.Column) {
					found = n
				}
			}
		}

		return true // keep digging
	})

	return found
}

func (h *Handler) findScopeAt(doc *Document, pos Position) *semantic.Scope {
	if doc.Info == nil {
		return nil
	}

	line := pos.Line + 1
	col := pos.Character + 1
	normFilename := normalizePath(uriToPath(doc.URI))

	var bestNode ast.Node
	var bestScope *semantic.Scope

	// Iterate over all nodes that have an associated scope
	for node, scope := range doc.Info.Scopes {
		start := node.Pos()
		// The scope must be in the same file
		if normalizePath(start.Filename) != normFilename {
			continue
		}
		// The scope must start before or at the cursor
		if start.Line < line || (start.Line == line && start.Column <= col) {
			// We want the "latest" such node (innermost)
			if bestNode == nil || start.Line > bestNode.Pos().Line || (start.Line == bestNode.Pos().Line && start.Column > bestNode.Pos().Column) {
				bestNode = node
				bestScope = scope
			}
		}
	}

	if bestScope != nil {
		return bestScope
	}

	// Fallback: use global scope if no inner scope found
	return nil
}

// makeFullRange creates a Range starting at the node's position and ending at the maximum position found in the node (or falling back to namePos + nameLen + 100).
func makeFullRange(node ast.Node, namePos token.Position, nameLen int) Range {
	startPos := node.Pos()
	startLine := startPos.Line - 1
	startChar := startPos.Column - 1
	if startLine < 0 {
		startLine = 0
	}
	if startChar < 0 {
		startChar = 0
	}

	maxPos := nodeMaxPosition(node)

	endLine := namePos.Line - 1
	endChar := namePos.Column - 1 + nameLen + 100 // default fallback

	if maxPos.Line > 0 {
		endLine = maxPos.Line - 1
		endChar = maxPos.Column - 1 + 200 // give a generous margin for characters on that line
	}

	if endLine < 0 {
		endLine = 0
	}
	if endChar < 0 {
		endChar = 0
	}

	// Safety adjustments to guarantee containment of the name
	if endLine < startLine {
		endLine = startLine
	}
	nameLine := namePos.Line - 1
	nameEndChar := namePos.Column - 1 + nameLen
	if nameLine < 0 {
		nameLine = 0
	}
	if nameEndChar < 0 {
		nameEndChar = 0
	}

	// Bulletproof containment:
	// 1. Ensure fullRange.Start <= selectionRange.Start
	if startLine > nameLine {
		startLine = nameLine
		startChar = namePos.Column - 1
		if startChar < 0 {
			startChar = 0
		}
	} else if startLine == nameLine && startChar > (namePos.Column-1) {
		startChar = namePos.Column - 1
		if startChar < 0 {
			startChar = 0
		}
	}

	// 2. Ensure fullRange.End >= selectionRange.End
	if endLine < nameLine {
		endLine = nameLine
		endChar = nameEndChar
	} else if endLine == nameLine && endChar < nameEndChar {
		endChar = nameEndChar
	}

	return Range{
		Start: Position{Line: startLine, Character: startChar},
		End:   Position{Line: endLine, Character: endChar},
	}
}

func nodeMaxPosition(node ast.Node) token.Position {
	maxPos := token.Position{}
	if node == nil || ast.IsNil(node) {
		return maxPos
	}
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil || ast.IsNil(n) {
			return true
		}
		pos := n.Pos()
		if pos.Line > 0 { // token.Position is valid if Line > 0
			if pos.Line > maxPos.Line {
				maxPos = pos
			} else if pos.Line == maxPos.Line && pos.Column > maxPos.Column {
				maxPos = pos
			}
		}
		return true
	})
	return maxPos
}

const semanticTokenString = 17
const semanticTokenOperator = 19

// addInterpolatedStringTokens highlights static string text and interpolation delimiters/operators.
// Expression identifiers inside ${...} are highlighted separately via semantic analysis.
func addInterpolatedStringTokens(content string, is *ast.InterpolatedString, addToken func(line, char, length, tokenType, tokenModifiers int)) {
	offset := is.Token.Position.Offset
	if offset < 0 || offset >= len(content) {
		return
	}
	totalLen := getStringLiteralLength(content, offset)
	if totalLen == 0 {
		return
	}

	innerStart := offset
	innerEnd := offset + totalLen
	switch {
	case innerEnd-innerStart >= 6 && content[innerStart:innerStart+3] == `"""`:
		innerStart += 3
		innerEnd -= 3
	case innerEnd > innerStart && content[innerStart] == '"':
		innerStart++
		innerEnd--
	case innerEnd > innerStart && content[innerStart] == '`':
		innerStart++
		innerEnd--
	}

	addAtOffset := func(off, length, tokenType int) {
		if length <= 0 || off < 0 || off+length > len(content) {
			return
		}
		pos := positionAtOffset(content, off)
		addToken(pos.Line-1, pos.Column-1, length, tokenType, 0)
	}

	for i := innerStart; i < innerEnd; {
		if content[i] == '$' {
			if i+1 < innerEnd && content[i+1] == '{' {
				addAtOffset(i, 2, semanticTokenOperator) // ${
				i += 2
				depth := 1
				for i < innerEnd && depth > 0 {
					if depth == 1 {
						switch content[i] {
						case '(', ')', '.', '[', ']', ',', '+', '-', '*', '/', '%':
							addAtOffset(i, 1, semanticTokenOperator)
						}
					}
					switch content[i] {
					case '{':
						depth++
					case '}':
						depth--
					}
					i++
				}
				if i > innerStart && content[i-1] == '}' {
					addAtOffset(i-1, 1, semanticTokenOperator) // }
				}
				continue
			}
			if i+1 < innerEnd && isInterpolationIdentStart(content[i+1]) {
				addAtOffset(i, 1, semanticTokenOperator) // $
				j := i + 1
				for j < innerEnd && isInterpolationIdentPart(content[j]) {
					j++
				}
				if j > i+1 {
					i = j
					continue
				}
			}
		}

		segStart := i
		for i < innerEnd {
			if content[i] == '$' && i+1 < innerEnd {
				next := content[i+1]
				if next == '{' || isInterpolationIdentStart(next) {
					break
				}
			}
			i++
		}
		if i > segStart {
			addAtOffset(segStart, i-segStart, semanticTokenString)
		}
	}
}

func identifierTokenLength(content string, ident *ast.Identifier) int {
	literal := ident.Token.Literal
	if literal == "" {
		literal = ident.Value
	}
	off := ident.Token.Position.Offset
	if off > 0 && off+len(literal) <= len(content) && content[off:off+len(literal)] == literal {
		return len(literal)
	}
	off = offsetAtLineColumn(content, ident.Token.Position.Line, ident.Token.Position.Column)
	if off >= 0 && off+len(literal) <= len(content) && content[off:off+len(literal)] == literal {
		return len(literal)
	}
	return len(literal)
}

func offsetAtLineColumn(content string, line, col int) int {
	if line < 1 || col < 1 {
		return -1
	}
	curLine, curCol := 1, 1
	for i := 0; i < len(content); i++ {
		if curLine == line && curCol == col {
			return i
		}
		if content[i] == '\n' {
			curLine++
			curCol = 1
		} else {
			curCol++
		}
	}
	return -1
}

func isInterpolationIdentStart(ch byte) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isInterpolationIdentPart(ch byte) bool {
	return isInterpolationIdentStart(ch) || (ch >= '0' && ch <= '9')
}

func positionAtOffset(content string, offset int) token.Position {
	line, col := 1, 1
	if offset > len(content) {
		offset = len(content)
	}
	for i := 0; i < offset; i++ {
		if content[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return token.Position{Line: line, Column: col, Offset: offset}
}

// getStringLiteralLength calculates the exact length of a string/rune literal in source text, including quotes and escapes.
func getStringLiteralLength(content string, offset int) int {
	if offset < 0 || offset >= len(content) {
		return 0
	}

	// Check for triple quotes
	if offset+3 <= len(content) && content[offset:offset+3] == `"""` {
		braceDepth := 0
		for i := offset + 3; i < len(content); i++ {
			if i+2 < len(content) && content[i] == '$' && content[i+1] == '{' {
				braceDepth++
				i++
				continue
			}
			if content[i] == '{' && braceDepth > 0 {
				braceDepth++
				continue
			}
			if content[i] == '}' && braceDepth > 0 {
				braceDepth--
				continue
			}
			if braceDepth == 0 && i+3 <= len(content) && content[i:i+3] == `"""` {
				return i + 3 - offset
			}
		}
		return len(content) - offset
	}

	quote := content[offset]
	if quote != '"' && quote != '`' && quote != '\'' {
		return 1
	}

	escaped := false
	braceDepth := 0
	for i := offset + 1; i < len(content); i++ {
		ch := content[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && quote != '`' {
			escaped = true
			continue
		}
		if ch == '$' && quote != '\'' && i+1 < len(content) && content[i+1] == '{' {
			braceDepth++
			i++
			continue
		}
		if ch == '{' && braceDepth > 0 {
			braceDepth++
			continue
		}
		if ch == '}' && braceDepth > 0 {
			braceDepth--
			continue
		}
		if ch == quote && braceDepth == 0 {
			return i + 1 - offset
		}
	}
	return len(content) - offset
}

// posToRange creates a Range from a token.Position and a length
func posToRange(pos token.Position, length int) Range {
	line := pos.Line - 1
	char := pos.Column - 1
	if line < 0 {
		line = 0
	}
	if char < 0 {
		char = 0
	}
	return Range{
		Start: Position{Line: line, Character: char},
		End:   Position{Line: line, Character: char + length},
	}
}

// deriveStdDir walks up the directory tree from the open document's path
// to locate the project's std/ folder.
func deriveStdDir(docURI string) string {
	// Strip file:// prefix and URL-decode basic characters
	docPath := strings.TrimPrefix(docURI, "file:///")
	docPath = strings.TrimPrefix(docPath, "file://")
	docPath = strings.ReplaceAll(docPath, "%20", " ")
	docPath = strings.ReplaceAll(docPath, "%3A", ":")
	docPath = strings.ReplaceAll(docPath, "%3a", ":")

	// Normalize slashes for Windows
	docPath = filepath.FromSlash(docPath)

	// 1. Check Env
	if env := os.Getenv("NORA_STD_PATH"); env != "" {
		return env
	}

	// 2. Check for local std or lib/std by walking up
	dir := filepath.Dir(docPath)
	for {
		// Try ./std
		candidate := filepath.Join(dir, "std")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		// Try ./lib/std
		candidate = filepath.Join(dir, "lib", "std")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		// Try ./libs/std
		candidate = filepath.Join(dir, "libs", "std")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// 3. Check relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		// Try ./std relative to exe
		path := filepath.Join(exeDir, "std")
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}
		// Try ../std relative to exe
		path = filepath.Join(exeDir, "..", "std")
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}
	}

	// 4. Check system locations (Unix-like)
	systemPaths := []string{
		"/usr/local/lib/Nora/std",
		"/lib/Nora/std",
		"/usr/lib/Nora/std",
	}
	for _, p := range systemPaths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}

	// 5. Final fallback: use cwd/std
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "std")
}

type Dependency struct {
	Path    string `yaml:"path"`
	Version string `yaml:"version"`
}

type ProjectConfig struct {
	Name         string                `yaml:"name"`
	Version      string                `yaml:"version"`
	Language     string                `yaml:"language"`
	Plugins      []string              `yaml:"plugins"`
	Dependencies map[string]Dependency `yaml:"dependencies"`
	NoStdlib     bool                  `yaml:"no_stdlib"`
	NoCore       bool                  `yaml:"no_core"`
}

type LSPFileLoader struct {
	Handler      *Handler
	Analyzer     *semantic.SemanticAnalyzer
	Program      *ast.Program
	StdDir       string
	Plugins      []string
	addedFiles   map[string]bool
	loadingPkgs  map[string]bool
	Dependencies map[string]Dependency
	NoStdlib     bool
	NoCore       bool
}

func findProjectRoot(docPath string) string {
	dir := filepath.Clean(docPath)
	if info, err := os.Stat(dir); err == nil && !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	for {
		manifestPath := filepath.Join(dir, "nora.yaml")
		if _, err := os.Stat(manifestPath); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func (f *LSPFileLoader) loadManifest(dirPath string) {
	manifestPath := filepath.Join(dirPath, "nora.yaml")
	if data, err := os.ReadFile(manifestPath); err == nil {
		var config ProjectConfig
		if err := yaml.Unmarshal(data, &config); err == nil {
			f.NoStdlib = config.NoStdlib
			f.NoCore = config.NoCore
			if f.Dependencies == nil {
				f.Dependencies = make(map[string]Dependency)
			}
			f.Plugins = append(f.Plugins, config.Plugins...)
			for name, dep := range config.Dependencies {
				if _, exists := f.Dependencies[name]; !exists {
					if !filepath.IsAbs(dep.Path) && !strings.HasPrefix(dep.Path, "http") {
						dep.Path = filepath.Join(dirPath, dep.Path)
					}
					f.Dependencies[name] = dep
				}
			}
		}
	}
}

func (f *LSPFileLoader) GetInstance(key string) (*ast.FunctionStatement, bool) {
	if val, ok := f.Handler.instanceCache.Load(key); ok {
		return val.(*ast.FunctionStatement), true
	}
	return nil, false
}

func (f *LSPFileLoader) SetInstance(key string, fn *ast.FunctionStatement) {
	f.Handler.instanceCache.Store(key, fn)
}

func (f *LSPFileLoader) findOpenDocument(path string) (*Document, bool) {
	normalizedPath := normalizePath(path)
	var foundDoc *Document
	f.Handler.docs.Range(func(key, value interface{}) bool {
		docURI := key.(string)
		docPath := uriToPath(docURI)
		if normalizePath(docPath) == normalizedPath {
			foundDoc = value.(*Document)
			return false // stop iteration
		}
		return true // continue iteration
	})
	return foundDoc, foundDoc != nil
}

func (f *LSPFileLoader) Load(importPath string) (*semantic.Scope, error) {
	if importPath == "" || importPath == "." || importPath == "./" {
		return nil, fmt.Errorf("invalid import path")
	}
	importPath = filepath.Clean(importPath)

	// 1. Check Dependencies from nora.yaml
	if dep, ok := f.Dependencies[importPath]; ok {
		actualPath := dep.Path
		// Load transitive dependencies
		libConfigPath := filepath.Join(actualPath, "nora.yaml")
		if data, err := os.ReadFile(libConfigPath); err == nil {
			var libConfig ProjectConfig
			if err := yaml.Unmarshal(data, &libConfig); err == nil {
				if f.Dependencies == nil {
					f.Dependencies = make(map[string]Dependency)
				}
				for transName, transDep := range libConfig.Dependencies {
					if _, exists := f.Dependencies[transName]; !exists {
						if !filepath.IsAbs(transDep.Path) && !strings.HasPrefix(transDep.Path, "http") {
							transDep.Path = filepath.Join(actualPath, transDep.Path)
						}
						f.Dependencies[transName] = transDep
					}
				}
			}
		}
		importPath = actualPath
	} else if !filepath.IsAbs(importPath) &&
		!strings.HasPrefix(importPath, "./") &&
		!strings.HasPrefix(importPath, "../") {
		// 2. Check if it exists in core/
		root := filepath.Dir(f.StdDir)
		coreCandidate := filepath.Join(root, "core", importPath)
		stdCandidate := filepath.Join(f.StdDir, importPath)
		
		if _, err := os.Stat(coreCandidate); err == nil {
			importPath = "core/" + importPath
		} else if _, err := os.Stat(coreCandidate + ".nr"); err == nil {
			importPath = "core/" + importPath
		} else if _, err := os.Stat(stdCandidate); err == nil {
			// 3. Check if it exists in std/
			importPath = "std/" + importPath
		} else if _, err := os.Stat(stdCandidate + ".nr"); err == nil {
			importPath = "std/" + importPath
		} else {
			// If not in core/ or std/, check if it exists locally, otherwise fallback to std/ just in case
			localCandidate := filepath.Join(root, importPath)
			if _, err := os.Stat(localCandidate); err != nil {
				if _, err := os.Stat(localCandidate + ".nr"); err != nil {
					importPath = "std/" + importPath
				}
			}
		}
	}

	// Determine full path
	fullPath := normalizePath(importPath)
	if !filepath.IsAbs(importPath) && f.StdDir != "" {
		// Project root is parent of std
		root := filepath.Dir(f.StdDir)
		candidate := filepath.Join(root, importPath)
		if _, err := os.Stat(candidate); err == nil {
			fullPath = normalizePath(candidate)
		} else {
			// Also try relative to std itself
			candidate2 := filepath.Join(f.StdDir, importPath)
			if _, err := os.Stat(candidate2); err == nil {
				fullPath = normalizePath(candidate2)
			}
		}
	}

	// Load package manifest if present to discover transitive dependencies
	if info, err := os.Stat(fullPath); err == nil {
		if info.IsDir() {
			f.loadManifest(fullPath)
		} else {
			f.loadManifest(filepath.Dir(fullPath))
		}
	}

	//fmt.Printf("[DEBUG] LSPFileLoader.Load: importPath = %q, fullPath = %q\n", importPath, fullPath)

	// 2. Check Global Package Cache
	if entry, ok := f.Handler.packageCache.Load(fullPath); ok {
		e := entry.(*packageCacheEntry)
		// Verify if any file in the package changed OR a new file was added
		changed := false
		if dirInfo, err := os.Stat(fullPath); err == nil {
			if dirInfo.ModTime() != e.DirModTime {
				changed = true
			}
		}

		if !changed {
			for path, oldTime := range e.ModTimes {
				if _, ok := f.findOpenDocument(path); ok {
					changed = true
					break
				}
				if info, err := os.Stat(path); err != nil || info.ModTime() != oldTime {
					changed = true
					break
				}
			}
		}

		// fmt.Printf("[DEBUG] Cache hit check for %q: hit = true, changed = %t\n", fullPath, changed)

		if !changed {
			// Cache hit! Reuse ASTs but CLONE/MERGE the Scope to avoid cross-thread mutation and scope orphaning
			var cloned *semantic.Scope
			if e.Scope.PackageName != "" {
				cloned = f.Analyzer.PackageScopes[e.Scope.PackageName]
			}
			if cloned == nil {
				cloned = semantic.NewScope(f.Analyzer.GlobalScope, e.Scope.Kind)
				cloned.PackageName = e.Scope.PackageName
				if cloned.PackageName != "" {
					f.Analyzer.PackageScopes[cloned.PackageName] = cloned
				}
			}
			for k, v := range e.CapturedSymbols {
				cloned.Symbols[k] = v
			}

			// Restore cached semantic metadata (MethodSymbols and FieldSymbols)
			for k, v := range e.MethodSymbols {
				if f.Analyzer.SemanticInfo.MethodSymbols == nil {
					f.Analyzer.SemanticInfo.MethodSymbols = make(map[types.NRType]map[string]*semantic.Symbol)
				}
				f.Analyzer.SemanticInfo.MethodSymbols[k] = v
			}
			for k, v := range e.FieldSymbols {
				if f.Analyzer.SemanticInfo.FieldSymbols == nil {
					f.Analyzer.SemanticInfo.FieldSymbols = make(map[*types.StructType]map[string]*semantic.Symbol)
				}
				f.Analyzer.SemanticInfo.FieldSymbols[k] = v
			}

			// Restore cached FuncScopes
			if f.Analyzer.FuncScopes == nil {
				f.Analyzer.FuncScopes = make(map[*ast.FunctionStatement]*semantic.Scope)
			}
			for k, v := range e.FuncScopes {
				scopeToRestore := v
				if v == e.Scope {
					scopeToRestore = cloned
				}
				f.Analyzer.FuncScopes[k] = scopeToRestore
			}

			// Restore cached Scopes
			for k, v := range e.Scopes {
				scopeToRestore := v
				if v == e.Scope {
					scopeToRestore = cloned
				}
				if f.Analyzer.SemanticInfo.Scopes == nil {
					f.Analyzer.SemanticInfo.Scopes = make(map[ast.Node]*semantic.Scope)
				}
				f.Analyzer.SemanticInfo.Scopes[k] = scopeToRestore
			}

			// Restore cached Defs
			for k, v := range e.Defs {
				if f.Analyzer.SemanticInfo.Defs == nil {
					f.Analyzer.SemanticInfo.Defs = make(map[*ast.Identifier]*semantic.Symbol)
				}
				f.Analyzer.SemanticInfo.Defs[k] = v
			}

			// Restore cached Uses
			for k, v := range e.Uses {
				if f.Analyzer.SemanticInfo.Uses == nil {
					f.Analyzer.SemanticInfo.Uses = make(map[*ast.Identifier]*semantic.Symbol)
				}
				f.Analyzer.SemanticInfo.Uses[k] = v
			}

			// Restore cached Types
			for k, v := range e.Types {
				if f.Analyzer.SemanticInfo.Types == nil {
					f.Analyzer.SemanticInfo.Types = make(map[ast.Node]types.NRType)
				}
				f.Analyzer.SemanticInfo.Types[k] = v
			}

			return cloned, nil
		}
	}

	// Circular dependency check!
	if f.loadingPkgs == nil {
		f.loadingPkgs = make(map[string]bool)
	}
	if f.loadingPkgs[fullPath] {
		// Return the scope early to break the cycle. It will be populated later.
		pkgName := filepath.Base(fullPath)
		return f.Analyzer.GetPackageScope(pkgName), nil
	}
	f.loadingPkgs[fullPath] = true
	defer delete(f.loadingPkgs, fullPath)

	// Record existing keys of all semantic maps to capture newly compiled ones.
	existingMethods := make(map[types.NRType]bool)
	for k := range f.Analyzer.SemanticInfo.MethodSymbols {
		existingMethods[k] = true
	}
	existingFields := make(map[*types.StructType]bool)
	for k := range f.Analyzer.SemanticInfo.FieldSymbols {
		existingFields[k] = true
	}
	existingFuncScopes := make(map[*ast.FunctionStatement]bool)
	for k := range f.Analyzer.FuncScopes {
		existingFuncScopes[k] = true
	}
	existingScopes := make(map[ast.Node]bool)
	for k := range f.Analyzer.SemanticInfo.Scopes {
		existingScopes[k] = true
	}
	existingDefs := make(map[*ast.Identifier]bool)
	for k := range f.Analyzer.SemanticInfo.Defs {
		existingDefs[k] = true
	}
	existingUses := make(map[*ast.Identifier]bool)
	for k := range f.Analyzer.SemanticInfo.Uses {
		existingUses[k] = true
	}
	existingTypes := make(map[ast.Node]bool)
	for k := range f.Analyzer.SemanticInfo.Types {
		existingTypes[k] = true
	}

	// Resolve to a concrete .nr file path if it's a directory
	if !strings.HasSuffix(fullPath, ".nr") {
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			// It's a directory, load all .nr files in it
			if strings.Contains(fullPath, "node_modules") || strings.Contains(fullPath, ".git") {
				return nil, fmt.Errorf("ignoring large directory")
			}
			files, err := os.ReadDir(fullPath)
			if err != nil {
				return nil, err
			}

			var pkgScope *semantic.Scope
			var NoraFiles []*ast.File
			modTimes := make(map[string]time.Time)

			for _, fileInfo := range files {
				if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".nr") {
					fullFilePath := filepath.Join(fullPath, fileInfo.Name())

					info, err := os.Stat(fullFilePath)
					if err != nil {
						continue
					}
					modTimes[fullFilePath] = info.ModTime()

					var existingFile *ast.File
					normPath := normalizePath(fullFilePath)
					for _, fNode := range f.Program.Files {
						if normalizePath(fNode.Name) == normPath {
							existingFile = fNode
							break
						}
					}

					var file *ast.File
					if existingFile != nil {
						file = existingFile
					} else {
						// Check if open in editor
						var input []byte
						if doc, ok := f.findOpenDocument(fullFilePath); ok {
							doc.mu.RLock()
							input = []byte(doc.Content)
							doc.mu.RUnlock()
						} else {
							input, err = os.ReadFile(fullFilePath)
							if err != nil {
								continue
							}
						}

						l := lexer.New(string(input), fullFilePath)
						l.Context = f.Analyzer.Context
						p := parser.New(l)
						p.AllowNoPackage = false
						p.Context = f.Analyzer.Context
						file = p.Parse(fullFilePath)

						if file.Name != "" && !f.addedFiles[file.Name] {
							f.Program.Files = append(f.Program.Files, file)
							f.addedFiles[file.Name] = true
						}
					}
					NoraFiles = append(NoraFiles, file)

					if pkgScope == nil {
						pkgName := f.Analyzer.GetPackageName(file)
						pkgScope = f.Analyzer.GetPackageScope(pkgName)
					}
				}
			}

			// Pass 1a: Collect Types ONLY for all files
			prevTypeCollection := f.Analyzer.TypeCollectionOnly
			f.Analyzer.TypeCollectionOnly = true
			for _, nf := range NoraFiles {
				f.Analyzer.ProcessedFiles[normalizePath(nf.Name)] = false
				f.Analyzer.CollectSymbols(nf)
			}

			// Pass 1b: Collect Imports, Functions, Methods for all files
			f.Analyzer.TypeCollectionOnly = false
			for _, nf := range NoraFiles {
				// Reset processed flag so the second pass actually runs
				f.Analyzer.ProcessedFiles[normalizePath(nf.Name)] = false
				f.Analyzer.CollectSymbols(nf)
			}
			f.Analyzer.TypeCollectionOnly = prevTypeCollection

			// Pass 1.5: populate struct fields, enum variants, interface methods
			// This must run AFTER all files are collected so cross-file type references resolve.
			if !f.Analyzer.TypeCollectionOnly {
				for _, nf := range NoraFiles {
					f.Analyzer.AnalyzeFileTypes(nf)
				}
			}

			if pkgScope != nil {
				dirModTime := time.Time{}
				if dirInfo, err := os.Stat(fullPath); err == nil {
					dirModTime = dirInfo.ModTime()
				}

				// Capture newly added semantic maps
				capturedMethods := make(map[types.NRType]map[string]*semantic.Symbol)
				for k, v := range f.Analyzer.SemanticInfo.MethodSymbols {
					if !existingMethods[k] {
						capturedMethods[k] = v
					}
				}
				capturedFields := make(map[*types.StructType]map[string]*semantic.Symbol)
				for k, v := range f.Analyzer.SemanticInfo.FieldSymbols {
					if !existingFields[k] {
						capturedFields[k] = v
					}
				}
				capturedFuncScopes := make(map[*ast.FunctionStatement]*semantic.Scope)
				for k, v := range f.Analyzer.FuncScopes {
					if !existingFuncScopes[k] {
						capturedFuncScopes[k] = v
					}
				}
				capturedScopes := make(map[ast.Node]*semantic.Scope)
				for k, v := range f.Analyzer.SemanticInfo.Scopes {
					if !existingScopes[k] {
						capturedScopes[k] = v
					}
				}
				capturedDefs := make(map[*ast.Identifier]*semantic.Symbol)
				for k, v := range f.Analyzer.SemanticInfo.Defs {
					if !existingDefs[k] {
						capturedDefs[k] = v
					}
				}
				capturedUses := make(map[*ast.Identifier]*semantic.Symbol)
				for k, v := range f.Analyzer.SemanticInfo.Uses {
					if !existingUses[k] {
						capturedUses[k] = v
					}
				}
				capturedTypes := make(map[ast.Node]types.NRType)
				for k, v := range f.Analyzer.SemanticInfo.Types {
					if !existingTypes[k] {
						capturedTypes[k] = v
					}
				}

				capturedSymbols := make(map[string]*semantic.Symbol)
				if pkgScope != nil {
					for k, v := range pkgScope.Symbols {
						definedHere := false
						if v.DefNode == nil {
							definedHere = true
						} else {
							defFilename := normalizePath(v.DefNode.Pos().Filename)
							for _, pf := range NoraFiles {
								if normalizePath(pf.Name) == defFilename {
									definedHere = true
									break
								}
							}
						}
						if definedHere {
							capturedSymbols[k] = v
						}
					}
				}

				f.Handler.packageCache.Store(fullPath, &packageCacheEntry{
					Scope:           pkgScope,
					CapturedSymbols: capturedSymbols,
					Files:           NoraFiles,
					ModTimes:      modTimes,
					DirModTime:    dirModTime,
					MethodSymbols: capturedMethods,
					FieldSymbols:  capturedFields,
					FuncScopes:    capturedFuncScopes,
					Scopes:        capturedScopes,
					Defs:          capturedDefs,
					Uses:          capturedUses,
					Types:         capturedTypes,
				})

				return pkgScope, nil
			}
		} else if !strings.HasSuffix(fullPath, ".nr") {
			fullPath += ".nr"
		}
	}

	// 3. Single-File Load (if not a directory)
	if strings.HasSuffix(fullPath, ".nr") {
		info, err := os.Stat(fullPath)
		if err != nil {
			return nil, err
		}

		var input []byte
		if doc, ok := f.findOpenDocument(fullPath); ok {
			doc.mu.RLock()
			input = []byte(doc.Content)
			doc.mu.RUnlock()
		} else {
			input, err = os.ReadFile(fullPath)
			if err != nil {
				return nil, err
			}
		}

		l := lexer.New(string(input), fullPath)
		l.Context = f.Analyzer.Context
		p := parser.New(l)
		p.AllowNoPackage = false
		p.Context = f.Analyzer.Context
		file := p.Parse(fullPath)

		if file.Name != "" && !f.addedFiles[file.Name] {
			f.Program.Files = append(f.Program.Files, file)
			f.addedFiles[file.Name] = true
		}

		// First pass: collect symbols
		f.Analyzer.CollectSymbols(file)
		// Pass 1.5: populate struct fields, enum variants, interface methods
		f.Analyzer.AnalyzeFileTypes(file)

		NoraFiles := []*ast.File{file}
		pkgScope := f.Analyzer.GetPackageScope(f.Analyzer.GetPackageName(file))

		dirModTime := time.Time{}
		if dirInfo, err := os.Stat(filepath.Dir(fullPath)); err == nil {
			dirModTime = dirInfo.ModTime()
		}

		// Capture newly added semantic maps
		capturedMethods := make(map[types.NRType]map[string]*semantic.Symbol)
		for k, v := range f.Analyzer.SemanticInfo.MethodSymbols {
			if !existingMethods[k] {
				capturedMethods[k] = v
			}
		}
		capturedFields := make(map[*types.StructType]map[string]*semantic.Symbol)
		for k, v := range f.Analyzer.SemanticInfo.FieldSymbols {
			if !existingFields[k] {
				capturedFields[k] = v
			}
		}
		capturedFuncScopes := make(map[*ast.FunctionStatement]*semantic.Scope)
		for k, v := range f.Analyzer.FuncScopes {
			if !existingFuncScopes[k] {
				capturedFuncScopes[k] = v
			}
		}
		capturedScopes := make(map[ast.Node]*semantic.Scope)
		for k, v := range f.Analyzer.SemanticInfo.Scopes {
			if !existingScopes[k] {
				capturedScopes[k] = v
			}
		}
		capturedDefs := make(map[*ast.Identifier]*semantic.Symbol)
		for k, v := range f.Analyzer.SemanticInfo.Defs {
			if !existingDefs[k] {
				capturedDefs[k] = v
			}
		}
		capturedUses := make(map[*ast.Identifier]*semantic.Symbol)
		for k, v := range f.Analyzer.SemanticInfo.Uses {
			if !existingUses[k] {
				capturedUses[k] = v
			}
		}
		capturedTypes := make(map[ast.Node]types.NRType)
		for k, v := range f.Analyzer.SemanticInfo.Types {
			if !existingTypes[k] {
				capturedTypes[k] = v
			}
		}

		// Save to Global Cache
		f.Handler.packageCache.Store(fullPath, &packageCacheEntry{
			Scope:         pkgScope,
			Files:         NoraFiles,
			ModTimes:      map[string]time.Time{fullPath: info.ModTime()},
			DirModTime:    dirModTime,
			MethodSymbols: capturedMethods,
			FieldSymbols:  capturedFields,
			FuncScopes:    capturedFuncScopes,
			Scopes:        capturedScopes,
			Defs:          capturedDefs,
			Uses:          capturedUses,
			Types:         capturedTypes,
		})

		return pkgScope, nil
	}

	return nil, fmt.Errorf("package not found: %s", importPath)
}

func uriToPath(uri string) string {
	if uri == "" {
		return ""
	}

	u, err := url.Parse(uri)
	if err == nil && u.Scheme == "file" {
		p := u.Path
		if runtime.GOOS == "windows" {
			// Windows file URIs can produce /C:/path; drop the leading slash.
			if strings.HasPrefix(p, "/") && len(p) > 2 && p[2] == ':' {
				p = strings.TrimPrefix(p, "/")
			}
		}

		decoded, err := url.PathUnescape(p)
		if err == nil {
			p = decoded
		}

		return normalizePath(filepath.FromSlash(p))
	}

	p := strings.TrimPrefix(uri, "file://")
	decoded, err := url.PathUnescape(p)
	if err == nil {
		p = decoded
	}

	return normalizePath(filepath.FromSlash(p))
}

func pathToURI(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "file://") {
		return path
	}

	// If the path is not absolute, try to resolve it relative to the current
	// working directory. This handles cases where AST/lexer stored a bare
	// filename (e.g. "main.nr") rather than an absolute path — common when
	// parsing in-memory or when callers passed a short name.
	resolved := path
	if !filepath.IsAbs(resolved) {
		cwd, err := os.Getwd()
		if err == nil {
			cand := filepath.Join(cwd, resolved)
			if _, err := os.Stat(cand); err == nil {
				resolved = cand
			}
		}
	}

	p := filepath.ToSlash(normalizePath(resolved))
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	// URL Encode spaces
	p = strings.ReplaceAll(p, " ", "%20")
	return "file://" + p
}

func normalizePath(path string) string {
	p := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		return strings.ToLower(p)
	}
	return p
}
