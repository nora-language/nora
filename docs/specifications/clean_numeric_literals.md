# Specification: Clean Numeric Literals & Automatic Promotion

## Title & Overview
Eliminating type suffixes from numeric literals in Nora, replacing them with automatic large integer promotion and contextual type coercion.

## Motivation
Historically, Nora supported numeric suffixes like `1000000000i64` or `50i64` to explicitly type literals. However, this introduced syntactic noise, complex suffix scanning in the lexer and parser, and ran contrary to Nora's design pillars: **User Simplicity** (clean, noise-free syntax) and **Language Consistency**.

Since Nora already supports standard constructor-style casts like `i64(...)` and implicit coercion of compatible integer types, explicit literal suffixes are redundant. By disallowing suffixes entirely, we achieve maximum code elegance and unify type conversions under the function-call syntax. To retain developer convenience, the compiler automatically promotes large integers (beyond 32-bit limits) to `i64` and contextually coerces literals in infix expressions to eliminate type mismatches.

## Syntax
Type suffixes (e.g. `i8`, `u8`, `i16`, `u16`, `i32`, `u32`, `i64`, `u64`, `f32`, `f64`, `c32`, `c64`, `c128`) are strictly disallowed on numeric literals.
- **Illegal**: `1000000000i64`, `10u8`, `9.8f32`
- **Legal**: `1000000000` (coerces to `i64` or `i32`), `i8(10)`, `f32(9.8)`
- **Imaginary literals**: The `i` and `j` suffixes remain allowed to represent complex component values (e.g. `5i`, `3.14i`).

## Semantics
1. **Parser Diagnostic**: If the parser extracts a disallowed type suffix, it emits a syntax error suggesting constructor-style casting and continues parsing.
2. **Automatic Promotion**: An unsuffixed integer literal is default-typed as `i32`. If the literal's parsed value is less than `-2147483648` or greater than `2147483647`, the semantic analyzer automatically types it as `i64`.
3. **Contextual Infix Promotion**: In a binary operation (such as `+`, `-`, `*`, `/`, `%`, `==`, `<`, etc.), if one operand is an unsuffixed integer literal and the other operand is a primitive integer type, the compiler automatically re-types the literal to match the other operand's type.

## Type Rules
```text
Γ ⊢ lit : IntegerLiteral
val(lit) < -2147483648  or  val(lit) > 2147483647
-------------------------------------------------
Γ ⊢ lit : i64 (Automatic Promotion)

Γ ⊢ op1 : T (where T is an integer type other than i32)
Γ ⊢ op2 : IntegerLiteral (unsuffixed, default i32)
-------------------------------------------------
Γ ⊢ op2 : T (Contextual Infix Promotion)
```

## Lease Rules
Coercions and promotions apply exclusively to copy-by-value primitive numeric types and literals, meaning they have no effect on ownership, moves, or lease safety.

## Examples
```nora
// 1. Large values (> 32-bit signed max) are automatically i64
var large = 6364136223846793005

// 2. Contextual type coercion in division/multiplication
var elapsed: i64 = 15000
var secs = elapsed / 1000 // '1000' is automatically promoted to i64

// 3. Explicit casting for other conversions
var byte_val = i8(10)
```

## Edge Cases
- **Shift Operations**: In shift operations (e.g. `raw >> 32`), the shift amount is contextually promoted to match the type of `raw`.
- **Negative Literals**: Negative values like `-2147483648` represent a unary minus prefix operator applied to `2147483648`. Since `2147483648` is promoted to `i64`, the prefix `-` operator is allowed on all numeric types to ensure this compiles flawlessly.

## Errors & Diagnostics
If a type suffix is used, the compiler yields a diagnostic error:
```text
type suffixes on numeric literals are disallowed; use constructor-style conversion (e.g. i64(value)) instead
```

## Future Considerations
We may explore further literal types (such as custom literal constructors) if Nora introduces user-defined literal parsing in the future.
