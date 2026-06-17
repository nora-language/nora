# String Interpolation

## Overview

Nora supports string interpolation, allowing developers to embed variables and expressions directly inside string literals without relying on cumbersome string concatenation or formatting utilities.

## Motivation

Creating dynamic strings via `+` operator concatenation or `printf`-style formatting functions is often error-prone and visually noisy. String interpolation provides a clean, highly readable syntax for building strings inline, leveraging the compiler to automatically resolve variables and types into their string representations.

## Syntax

Interpolation uses the `${expression}` syntax inside any standard string literal (double quotes `"`).

### 1. Basic Variable Injection

You can inject primitive variables directly into a string.

```nora
var name = "Nora"
var version = 1
var greeting = "Hello, ${name}! Version: ${version}"
io.PrintLn(greeting)
```

### 2. Expression Evaluation

The brackets can contain any valid Nora expression. The compiler evaluates the expression and converts the result to a string.

```nora
var a = 5
var b = 10
var calc = "Result: ${a + b} is greater than ${a}"
io.PrintLn(calc) // Result: 15 is greater than 5
```

### 3. Boolean Conversion

Primitive boolean types are automatically converted to "true" or "false".

```nora
var flag = true
var status = "Status is ${flag}"
```

## Semantics

1.  **Lexing & Parsing:** The lexer specifically identifies `${...}` patterns inside string tokens (`pkg/lexer` and `pkg/lsp/handler.go` logic) and passes them to the parser as distinct syntax nodes (`ast.InterpolatedString`).
2.  **Conversion:** Under the hood, the compiler evaluates the expression inside `${}`. For primitive types (integers, booleans), it implicitly calls standard library string-conversion routines before concatenating them together.

## Examples

### Complex Expressions
Interpolation can invoke functions or access struct fields seamlessly.

```nora
type User = struct { name: str }
fn (u: #User) get_id() i32 { return 99 }

var u = User { name: "Alice" }
var summary = "User ${u.name} has ID ${u.get_id()}"
```

## Errors & Diagnostics

*   **Invalid Expression:** Placing an invalid expression or an un-stringifiable complex type inside `${...}` will result in a compile-time error.
*   **Missing Closing Brace:** Forgetting the `}` on an interpolation block will trigger a lexer/parser error regarding an unterminated string literal.
