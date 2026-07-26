# Implementation Plan: Implicit Function Overloading

**Status**: Proposed
**Metadata**: Author: AI Agent
**Goal**: Implement ad-hoc function and method overloading implicitly. The compiler will automatically group functions with the same name into an overload set, provided their parameter types differ.

## Architecture & Philosophy
By removing the explicit attribute requirement, Nora achieves **User Simplicity**. Mathematical libraries and APIs become much cleaner to write and read, mimicking the behavior of C++ and Java. The compiler will silently handle the complexity of grouping colliding names, only throwing a "duplicate symbol" error if the exact same signature is declared twice.

## Affected Compiler Components

### 1. `pkg/semantic/model.go` (Symbol Scope)
* **Action**: Introduce a new `SymbolKind`, e.g., `SymOverloadGroup`.
* **Action**: Update `Scope.Define`. When a name collision occurs:
  - If the existing symbol is a `SymFunc`, convert it into a `SymOverloadGroup` and add both functions to it.
  - If the existing symbol is already a `SymOverloadGroup`, add the new function to the group.
  - Throw an error ONLY if the new function has the exact same parameter types as an existing one in the group.

### 2. `pkg/semantic/analyzer.go` (Semantic Analysis)
* **Action**: Update `CollectSymbols` (Pass 1). 
    * Mangle the internal C names of overloaded functions based on their parameter types (e.g., `process__i32`, `process__str`) so C11 codegen doesn't fail.
* **Action**: Update `AnalyzeCallExpression` (Pass 2). 
    * When a function call resolves to a `SymOverloadGroup`, recursively type-check all arguments first.
    * Implement `resolveOverload(group, argTypes)` to select the function signature that matches the arguments.
    * Save the resolved target symbol in `sa.SemanticInfo.Defs`.

### 3. `pkg/codegen/hir_codegen.go` (C11 Codegen)
* **Action**: Ensure that overloaded functions are emitted using their type-mangled names. 
* **Action**: Ensure call sites emit the mangled name of the matched overload.

## Risks & Edge Cases
* **Ambiguity**: If `fn(x: int)` and `fn(x: i32)` are defined, passing a literal `5` could match both. The compiler must have strict rules for type coercion prioritization or throw an "ambiguous call" diagnostic.

## Completion Criteria
* `Nora test` passes all existing tests.
* Positive tests run correctly, dispatching to the right overloaded function.
* Negative tests correctly emit diagnostics for ambiguous calls or missing overloads.
