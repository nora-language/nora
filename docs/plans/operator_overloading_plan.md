# Implementation Plan: Explicit Operator Overloading (`operator+`)

**Status**: Proposed
**Metadata**: Author: AI Agent
**Goal**: Transition Nora from relying on hardcoded duck-typing method names (e.g., `add`, `sub`) to explicit operator syntax (e.g., `operator+`, `operator-`). This prevents namespace pollution and accidental overloading.

## 1. Syntax Design
Operators will be defined as methods on structs using the `operator` keyword followed by the symbol:
```nora
pub fn (self: &Vector) operator+(other: Vector) Vector { ... }
pub fn (self: &Vector) operator==(other: Vector) bool { ... }
```

## 2. Affected Compiler Components

### A. Lexer & Parser (`pkg/parser/parser.go`)
* **Action**: Update `parseFunctionStatement`.
* When parsing the function's name (which expects an `IDENTIFIER`), check if the identifier's literal value is `"operator"`.
* If it is `"operator"`, look at the *next* token (`p.peekToken`). 
* If the next token is an operator token (`+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `>`, etc.), consume it and concatenate the literal into a single `ast.Identifier` (e.g., `"operator+"`).
* Note: This keeps `"operator"` as a normal identifier if used elsewhere, avoiding the need for a new hard keyword.

### B. Semantic Analyzer (`pkg/semantic/analyzer.go`)
* **Action**: Update operator resolution in `analyzeBinaryExpression` / `analyzeInfixExpression`.
* Currently, `case "+": methodName = "add"`.
* Change this to `case "+": methodName = "operator+"`.
* Apply this mapping to all supported operators (`-` -> `operator-`, `*` -> `operator*`, `==` -> `operator==`).

### C. C11 Codegen (`pkg/codegen/hir_codegen.go`)
* **Action**: Ensure that methods named `operator+` are safely mangled into valid C identifiers.
* E.g., `Vector_operator+` -> `Vector_op_add`.
* Add a simple mangling map for operator symbols to string representations before generating the C code.

## 3. Migration
* Update existing tests (like `operator_overload_test.nr`) that use `add`, `sub`, and `eq` to use the new `operator+`, `operator-`, and `operator==` syntax.

## 4. Completion Criteria
* The parser successfully parses `operator+` as a single function name.
* `vector + scalar` implicitly calls the correctly typed `operator+` method (thanks to implicit function overloading).
* Existing codebase operators are refactored, and all tests pass.
