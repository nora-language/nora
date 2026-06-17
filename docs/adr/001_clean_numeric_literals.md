# ADR 001: Clean Numeric Literals & Automatic Promotion

## Status
Accepted

## Context
Nora originally supported explicit type suffixes on numeric literals (e.g. `1000000000i64`, `50i64`, `10u8`). While functional, suffixes added syntactic noise, increased parsing complexity, and introduced potential ambiguities (e.g., parsing hex literals like `0x123f64` where `f64` could be a hex character or a float suffix).

Nora already features:
1. Constructor-style casting: `i64(50)` or `i8(10)`.
2. Safe integer widening/narrowing coercions in assignments and comparisons (`IsAssignable`).

Therefore, literal suffixes are structurally redundant and clash with our goals of **User Simplicity** (clean, noise-free syntax) and **Language Consistency**.

## Decision
Disallow all numeric type suffixes (like `i64`, `u64`, `f32`, `f64`, etc.) on literals. Instead, we implement two core compiler mechanisms to maximize developer convenience:
1. **Automatic Large Integer Promotion**: Unsuffixed integer literals exceeding 32-bit signed range limits (`[-2147483648, 2147483647]`) automatically type-check as `i64` in the semantic analyzer.
2. **Contextual Infix Promotion**: Unsuffixed integer literals involved in binary operations automatically scale/promote to match the type of their sibling operand, preventing type mismatches in everyday mathematical operations.
3. **Explicit Casts**: Standard constructor syntax like `i64(...)` or `i8(...)` is utilized for any other explicit conversion scenarios.

## Alternatives Considered
1. **Keep Suffixes**: Rust and C++ both feature literal suffixes. However, Nora values aesthetic cleanliness, minimal parsing overhead, and single-source type declarations far more than historical precedent.
2. **Require Explicit Casts Everywhere**: Disallowing suffixes and requiring constructors like `i64(1000000000)` or `i64(60)` for all non-i32 operations. This option was rejected as too verbose for large constants or basic mathematical operations.

## Consequences
- **Syntax Aesthetics**: Code written in Nora is cleaner and free of noisy suffixes.
- **Parsing Cleanliness**: Simplifies suffix scanning in `pkg/lexer` and `pkg/parser`.
- **Diagnostic DX**: Using a suffix yields a helpful compiler syntax error instructing the user to transition to constructor casts.
- **Backward Compatibility**: Any existing standard library files and integration tests utilizing type suffixes have been refactored to conform to the new design.
