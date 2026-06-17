# Investigation Report: User-Defined Literals in Nora

## Overview
This report evaluates the feasibility, language design impact, and implementation strategies for supporting user-defined literals (such as `10km`, `500ms`, `4GB`) in the Nora Programming Language.

## Background Context
Nora recently adopted the [Clean Numeric Literals Specification](file:///e:/Project/Project%20Chronos/second/docs/specifications/clean_numeric_literals.md), which strictly disallows type suffixes (e.g. `i64`, `f32`) on numeric literals in favor of constructor-style conversions:
* **Illegal**: `1000i64`
* **Legal**: `i64(1000)`

By standardizing on constructor-style casts, Nora simplifies parser logic and maintains a clean syntax. Introducing user-defined suffix literals (e.g., `10km`) requires careful evaluation against Nora's core pillars of **User Simplicity**, **Language Consistency**, and **Compiler Maintainability**.

---

## Analyzed Solutions

We evaluated three potential designs to achieve user-defined unit literals without baking specific units into the compiler's core runtime.

### Option 1: Syntactic Suffix Parsing (C++ Style)
In this model, the lexer and parser are modified to recognize any identifier immediately following a number as a user-defined literal token. The compiler then desugars `10km` into a function call.

#### Code Example
```nora
// User definition
pub fn Suffix_km(value: i64) @Distance {
    return alloc Distance { meters: value * 1000 }
}

// Usage
var dist = 10km // Desugars to Suffix_km(10)
```

#### Feasibility and Impact
* **Lexer/Parser Complexity**: High. The lexer must distinguish between valid hexadecimal integers (e.g., `0x10f`) and standard decimal integers with suffixes starting with `f`. Backtracking or parser-level token stitching is required.
* **Namespace Pollution & Resolution**: The parser must parse the suffix as an identifier, but full name resolution happens in a later compiler phase (`pkg/symbol_scope`). This creates a circular dependency where parsing structures depend on symbol resolution.
* **Language Consistency**: Poor. This reintroduces the exact suffix-scanning complexity that the Clean Numeric Literals specification removed.

---

### Option 2: Literal Conversion Interfaces (Swift/Rust Style)
Instead of custom suffix markers, custom types implement compiler-known interfaces such as `std.ExpressibleByIntegerLiteral`. When the compiler encounters an integer literal in a context expecting a custom type, it automatically invokes the constructor.

#### Code Example
```nora
pub type Distance = struct {
    meters: i64
}

// Automatically called when an integer literal is assigned to a Distance variable
pub fn (d: &Distance) FromIntegerLiteral(value: i64) {
    d.meters = value
}

// Usage
var dist: Distance = 10000 // Automatically converted
```

#### Feasibility and Impact
* **Aesthetics**: Clean. Avoids suffixes altogether, maintaining a minimal syntactic footprint.
* **Limitations**: Does not support unit suffixes. It cannot distinguish between `10km` and `10m` since both are just integer literals.

---

### Option 3: Constructor-Style Casts & Method Chaining (Nora Native Style - RECOMMENDED)
This option uses standard language constructs (functions or method calls) to convert raw numeric values into domain-specific types.

#### Code Example
```nora
// 1. Constructor-style functions (Aligned with i64(x))
pub fn km(value: i64) @Distance {
    return alloc Distance { meters: value * 1000 }
}

pub fn ms(value: i64) @Duration {
    return alloc Duration { milliseconds: value }
}

pub fn GB(value: i64) @ByteSize {
    return alloc ByteSize { bytes: value * 1024 * 1024 * 1024 }
}

// Usage
var dist = km(10)
var delay = ms(500)
var memory = GB(4)
```

#### Feasibility and Impact
* **Feasibility**: 100% supported by Nora's existing compiler architecture.
* **Language Consistency**: Perfect. It unifies all unit and type conversions under the standard function-call or constructor-style syntax. This matches how developers already convert types (e.g., `i64(5)`).
* **Predictability**: High. Every allocation and type conversion is explicit, leaving no hidden compiler magic or syntactic surprises.
* **Performance**: Excellent. Functions are easily inlined by the compiler's code generator or optimization passes.

---

## Conclusion & Recommendation

**Option 3 (Constructor-Style Casts)** is the best solution for Nora. 

Implementing custom suffix parsing (Option 1) introduces high parser complexity, violates the phase separation between parsing and name resolution, and contradicts the decisions of the Clean Numeric Literals specification. 

By utilizing standard constructor functions (e.g., `km(10)`), Nora preserves its clean, suffix-free numeric syntax while offering a type-safe, highly readable, and zero-overhead mechanism for defining domain-specific units.
