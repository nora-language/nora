# Language Server Protocol (LSP)

## Overview

The Nora compiler ships with a fully integrated, zero-configuration Language Server Protocol (LSP) implementation. It is invoked via the `Nora lsp` command and communicates over standard input/output (`stdio`), allowing seamless integration into major code editors like VS Code, Neovim, and IntelliJ.

## Implemented Capabilities

The `pkg/lsp/server.go` implementation registers and responds to a wide array of JSON-RPC requests, providing real-time compiler feedback.

### 1. Diagnostics & Real-time Errors
As files are opened (`textDocument/didOpen`) and modified (`textDocument/didChange`), the LSP continuously runs the lexical, parsing, semantic, and topological lease solving passes. Syntax errors and lifetime violations are instantly reported as standard diagnostics.

### 2. Hover Documentation (`textDocument/hover`)
Hovering over variables, function names, or structs provides semantic type information, signature details, and associated docstrings (if available).

### 3. Go to Definition (`textDocument/definition`)
Users can `Ctrl+Click` (or equivalent) to jump directly to the source declaration of a variable, function, struct, or module import across the entire workspace.

### 4. Find References (`textDocument/references`)
Retrieves all locations in the workspace where a specific symbol (variable, function, or type) is referenced.

### 5. Semantic Tokens (`textDocument/semanticTokens/full`)
Provides deep, compiler-accurate syntax highlighting. Rather than relying on inaccurate regex-based TextMate grammars, the editor colorizes tokens based on their exact AST resolution (e.g., distinguishing between a type, a local variable, and a keyword).

### 6. Autocomplete & Suggestions (`textDocument/completion`)
Context-aware code completion for variables in scope, struct fields, and accessible package functions.

### 7. Signature Help (`textDocument/signatureHelp`)
While typing inside function call parentheses `()`, the LSP displays a pop-up containing the expected parameter types and names, highlighting the active parameter being typed.

### 8. Document Formatting (`textDocument/formatting`)
Delegates to `pkg/format` to reformat the current document according to Nora's standard rules (or the workspace's `nora_fmt.yaml` config) when the user triggers a save or manual format command.

### 9. Renaming (`textDocument/rename` & `textDocument/prepareRename`)
Safely performs workspace-wide renaming of symbols. The `prepareRename` capability verifies that the symbol under the cursor is legally mutable before allowing the rename prompt.

### 10. Code Actions (`textDocument/codeAction`)
Provides contextual quick-fixes for common diagnostics (e.g., auto-importing a missing package or suggesting a missing field in a struct initialization).

### 11. Inlay Hints (`textDocument/inlayHint`)
Injects ghost-text directly into the editor for implicitly inferred types and parameter names, significantly improving readability of code relying on type inference (`var x = 10` -> `var x /* : i32 */ = 10`).

### 12. Document Symbols (`textDocument/documentSymbol`)
Provides an outline of the current file (functions, structs, variables), allowing editors to render a breadcrumb navigation or an interactive file outline tree.
