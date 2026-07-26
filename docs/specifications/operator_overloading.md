# Operator Overloading Specification

## Overview
Nora supports formalized operator overloading via explicit method implementations on `struct` types. Instead of relying on implicit, ad-hoc method names (e.g., `add()`), Nora enforces an explicit `operator<op>` naming scheme (e.g., `operator+()`).

## Motivation
To provide ergonomic syntax for custom data structures (like `Vector2D`, `Matrix`, `ComplexNumber`) without requiring manual method calls (e.g., `v1.add(v2)`). Formalizing these operations with the explicit `operator` prefix ensures clarity, predictability, and prevents unintended collisions with regular struct methods. This integrates seamlessly into Nora's RAII Topological Lease Solver and C11 Monomorphization backend.

## Syntax and Mapping
Operator overloads are defined by implementing specific methods prefixed with `operator` on a type. The compiler automatically maps binary and unary operators to these methods:

| Operator | Method Name | Example Translation |
|----------|-------------|---------------------|
| `+`      | `operator+` | `v1 + v2` ➔ `v1.operator+(v2)` |
| `-`      | `operator-` | `v1 - v2` ➔ `v1.operator-(v2)` |
| `*`      | `operator*` | `v1 * v2` ➔ `v1.operator*(v2)` |
| `/`      | `operator/` | `v1 / v2` ➔ `v1.operator/(v2)` |
| `%`      | `operator%` | `v1 % v2` ➔ `v1.operator%(v2)` |
| `==`, `!=` | `operator==` | `v1 == v2` ➔ `v1.operator==(v2)` (and inverted for `!=`) |
| `<`, `<=`, `>`, `>=` | `operator<`, etc. | `v1 < v2` ➔ `v1.operator<(v2)` |
| `[]`     | `operator[]` | `v[k]` ➔ `v.operator[](k)` (for reading) |
| `[]=`    | `operator[]_mut` | `v[k] = val` ➔ `*v.operator[]_mut(k) = val` (for writing) |
| `&`      | `operator&` | `v1 & v2` ➔ `v1.operator&(v2)` |
| `|`      | `operator\|` | `v1 \| v2` ➔ `v1.operator\|(v2)` |
| `^`      | `operator^` | `v1 ^ v2` ➔ `v1.operator^(v2)` |
| `<<`     | `operator<<`| `v1 << v2` ➔ `v1.operator<<(v2)` |
| `>>`     | `operator>>`| `v1 >> v2` ➔ `v1.operator>>(v2)` |

## Semantics and Type Rules

1. **Method Signatures**: 
   The overloaded methods must match the expected signatures. Typically, the receiver should be a leased reference (`#Self`) to avoid unnecessary consumption, though owned receivers (`@Self`) are permitted if the operation logically consumes the left-hand operand.
   - Example Arithmetic: `pub fn (self: #Vector2D) operator+(other: Vector2D) Vector2D`
   - Example Comparison: `pub fn (self: #Vector2D) operator==(other: #Vector2D) bool`

2. **AST to HIR Lowering**:
   If an operator involves a `struct` (whether leased or owned) on the left-hand side, the Semantic Analyzer correctly routes the type checking to the respective `operator` method. The HIR Lowerer (`pkg/hir/lower.go`) prevents it from decaying into a raw C binary operation (`hir.BinOp`). Instead, the AST Codegen backend emits the correctly mangled C-level function call (e.g., `Vector2D_operator_plus(NULL, &v1, v2)`).

3. **Fallback to Primitives**:
   If the type is a primitive, or if the type is a struct but no matching `operator` method is found, the compiler falls back to primitive C arithmetic (and emits a semantic error if applied to a struct without overloads).

## Examples

```nora
pub type Vector2D = struct {
    x: i32
    y: i32
}

pub fn (self: #Vector2D) operator+(other: Vector2D) Vector2D {
    return Vector2D{
        x: self.x + other.x,
        y: self.y + other.y,
    }
}

pub fn (self: #Vector2D) operator==(other: #Vector2D) bool {
    return self.x == other.x && self.y == other.y
}

pub fn main() {
    var v1 = Vector2D{x: 10, y: 20}
    var v2 = Vector2D{x: 5, y: 15}

    // Resolves to: Vector2D_operator_plus(&v1, v2)
    var v3 = v1 + v2

    // Resolves to: Vector2D_operator_eq(&v1, &v2)
    if v1 == v2 {
        panic("Equality failed")
    }
}
```

## Edge Cases
- **Pointer Arithmetic**: Pointer arithmetic remains unsupported by default to uphold memory safety guarantees. If `+` is used on a pointer, it triggers a semantic error unless it is a leased struct type implementing the `operator+` method.
- **Order of Operations**: Operator overloading does not change standard operator precedence.
- **Short-Circuiting**: Comparison and logical operators preserve standard short-circuiting semantics.
- **Operator Mangling**: The compiler handles translation of operator characters into C-compatible safe identifiers (e.g. `+` becomes `_plus`, `==` becomes `_eq`) behind the scenes.

## Future Considerations
- More complex custom desugaring for iterator loops via `in` operators.
