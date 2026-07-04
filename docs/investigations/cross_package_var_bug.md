# Investigation: Semantic Analyzer Cross-Package Global Variable Bug

**Status:** Open  
**Date:** 2026-07-04  
**Component:** `pkg/semantic/analyzer.go`

## The Problem
When declaring a public global variable (`pub var`) in one package and attempting to read it from another package, the compiler crashes with a semantic type mismatch error: `got void`.

For example, given `pkg_a.nr`:
```nora
package pkg_a
pub type Config = struct { val: i32 }
pub var GlobalConfig: Config = Config { val: 5 }
```

And `main.nr`:
```nora
package main
import "pkg_a"
fn main() {
    var c = pkg_a.GlobalConfig
}
```

Compiling `main.nr` throws:
```
Error: type mismatch: expected Config, got void
```

## Root Cause Analysis
The bug occurs due to the lack of topological sorting for cross-package `ast.VarStatement` evaluation during the Semantic Analyzer's `Pass 2` (body analysis).

1. **Pass 1 (Symbol Collection)**: In `analyzer.go`, when an `ast.VarStatement` is collected, its symbol is unconditionally initialized to `types.Void`:
   ```go
   sym, err := sa.CurrentScope.Define(n.Name.Value, types.Void, SymVar, n)
   ```
2. **Pass 1.5 (Type Analysis)**: Type definitions (structs, enums) and Function return types are resolved successfully across all packages.
3. **Pass 2 (Body Analysis)**: The analyzer iterates through the AST files in a flat order based on how they were parsed (usually `main.nr` first, followed by its dependencies).
   - When the analyzer evaluates `main.nr`, it encounters the `ast.SelectorExpression` `pkg_a.GlobalConfig`.
   - It looks up the symbol in `pkg_a`'s `ModuleType.Exports` scope.
   - Because `pkg_a.nr` has not yet been processed by `Analyze()` in Pass 2, the `sym.Type` for `GlobalConfig` is still stuck at its default `types.Void`.
   - `main.nr` crashes.

If `GlobalConfig` were declared in the *same file* or *same package*, it would occasionally work due to localized file evaluation order, masking the bug.

## Recommended Fix
The `SemanticAnalyzer` must introduce a topological sorting mechanism for `VarStatement` evaluation, similar to how it handles `TypeStatement`s in Pass 1.5.

Before executing Pass 2 (Body Analysis), the compiler should execute a pre-pass that evaluates the Right-Hand Side (RHS) of all top-level `ast.VarStatement`s globally and updates their `sym.Type`. This ensures that all global variable types are finalized before any function bodies or cross-package selector expressions attempt to read them.
