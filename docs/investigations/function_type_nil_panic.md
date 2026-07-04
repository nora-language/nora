# Investigation: Function Type Alias Nil Pointer Panic

## Status
Open

## Problem
The Nora compiler crashes with a `panic: runtime error: invalid memory address or nil pointer dereference` in `pkg/semantic/analyzer.go` when parsing a function type alias that contains named parameters (which is invalid syntax in Nora).

## Reproduction
The crash can be reproduced by attempting to compile a file containing a function type alias with named parameters:

```nora
pub type LogCallback = fn(level: i32, msg: StringView, userdata: ptr)
```

**Crash Trace**:
```text
panic: runtime error: invalid memory address or nil pointer dereference
[signal 0xc0000005 code=0x0 addr=0x18 pc=0x7ff6cd7900df]

goroutine 1 [running]:
github.com/nora-language/nora/pkg/semantic.(*SemanticAnalyzer).Analyze(0x28406c8423c0, {0x7ff6cddec970, 0x28406ca19ae0})
	E:/Project/Project Nora/nora/pkg/semantic/analyzer.go:1877 +0xd31f
```

## Root Cause
The panic is the result of a two-step failure between the Parser and the Semantic Analyzer:

1. **Parser Failure (`pkg/parser/parser.go`)**: 
   When parsing the right-hand side of the alias (`fn(level: i32)`), the `parseFunctionType` function expects type nodes separated by commas. When it parses the identifier `level`, the next token is a colon `:`. Because it expects a comma `,` or a closing parenthesis `)`, it fails to parse the parameter list. In this failure state, `expectPeek(token.RPAREN)` evaluates to false, and the function returns `nil`.
   Consequently, the `TypeStatement` AST node is created with its `Value` field set to `nil`.

2. **Semantic Analyzer Nil Dereference (`pkg/semantic/analyzer.go`)**:
   During Pass 2 of semantic analysis, the `Analyze` method attempts to process the `TypeStatement`. Because `n.Value` is `nil`, the type assertion `n.Value.(ast.TypeNode)` fails. The execution falls into the `else` block:
   ```go
   if tn, ok := n.Value.(ast.TypeNode); ok {
       resolvedType := sa.resolveTypeNode(tn)
       sym.Type = resolvedType
   } else {
       sa.AddError(n.Value.Pos(), "expected struct, interface, enum definition, or type alias")
       // ...
   }
   ```
   The `else` block calls `n.Value.Pos()`. Since `n.Value` is `nil`, this results in a nil pointer dereference and crashes the compiler.

## Fix
1. **Semantic Analyzer Fortification**:
   Update `analyzer.go` to explicitly check if `n.Value == nil` before attempting to call `n.Value.Pos()`. If it is `nil`, it should fallback to using `n.Name.Pos()` or `n.Token.Position` for the diagnostic error.
   
2. **Parser Recovery**:
   Modify `parseFunctionType` in `parser.go` so that if it encounters an invalid token, it does not return `nil`. Instead, it should emit a diagnostic error (e.g., "function types cannot have parameter names") and return a placeholder `ast.ErrorNode` to keep the AST intact and prevent downstream nil pointers.

## Validation
To validate the fix, a negative test should be added to `pkg/cmd/test/` (e.g., `fail_named_parameters_in_type.nr`) ensuring that compiling the reproductive snippet emits a clean diagnostic error rather than panicking.
