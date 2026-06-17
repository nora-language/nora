# Operator Overloading Specification

## Overview
Nora supports formalized operator overloading via special method implementations on `struct` types. Instead of relying on hardcoded language intrinsics or falling back to raw C primitive operations, Nora resolves arithmetic and comparison operators into method calls if a corresponding method exists. 

## Motivation
To provide ergonomic syntax for custom data structures (like `Vector2D`, `Matrix`, `ComplexNumber`) without requiring manual method calls (e.g., `v1.add(v2)`). Formalizing these operations inside Nora's `pkg/semantic` and `pkg/hir` pipeline ensures they interact correctly with the RAII Topological Lease Solver and C11 Monomorphization backend.

## Syntax and Mapping
Operator overloads are defined by implementing specific methods on a type. The compiler automatically maps binary operators to these methods:

| Operator | Method Name | Example Translation |
|----------|-------------|---------------------|
| `+`      | `add`       | `v1 + v2` ➔ `v1.add(v2)` |
| `-`      | `sub`       | `v1 - v2` ➔ `v1.sub(v2)` |
| `*`      | `mul`       | `v1 * v2` ➔ `v1.mul(v2)` |
| `/`      | `div`       | `v1 / v2` ➔ `v1.div(v2)` |
| `%`      | `mod`       | `v1 % v2` ➔ `v1.mod(v2)` |
| `==`, `!=` | `eq`      | `v1 == v2` ➔ `v1.eq(v2)` |
| `<`, `<=`, `>`, `>=` | `cmp` | `v1 < v2` ➔ `v1.cmp(v2) < 0` |
| `-` (Unary)| `neg`       | `-v` ➔ `v.neg()` |
| `!` (Unary)| `not`       | `!v` ➔ `v.not()` |
| `~` (Unary)| `bitnot`    | `~v` ➔ `v.bitnot()` |
| `[]`     | `index`     | `v[k]` ➔ `v.index(k)` |
| `[]=`    | `index_mut` | `v[k] = val` ➔ `*v.index_mut(k) = val` |
| `&`      | `bitand`    | `v1 & v2` ➔ `v1.bitand(v2)` |
| `|`      | `bitor`     | `v1 \| v2` ➔ `v1.bitor(v2)` |
| `^`      | `bitxor`    | `v1 ^ v2` ➔ `v1.bitxor(v2)` |
| `<<`     | `shl`       | `v1 << v2` ➔ `v1.shl(v2)` |
| `>>`     | `shr`       | `v1 >> v2` ➔ `v1.shr(v2)` |

## Semantics and Type Rules

1. **Method Signatures**: 
   The overloaded methods must match the expected signatures. Typically, the receiver should be a leased reference (`#Self`) to avoid unnecessary consumption, though owned receivers (`@Self`) are permitted if the operation logically consumes the left-hand operand.
   - Example Arithmetic: `pub fn (self: #Vector2D) add(other: Vector2D) Vector2D`
   - Example Comparison: `pub fn (self: #Vector2D) eq(other: #Vector2D) bool`

2. **AST to HIR Lowering**:
   If an operator involves a `struct` (whether leased or owned) on the left-hand side, the HIR Lowerer (`pkg/hir/lower.go`) prevents it from decaying into a raw C binary operation (`hir.BinOp`). Instead, it wraps it in an `ASTExpr` so the AST Codegen backend (`pkg/codegen/expressions.go`) can emit the correct C-level function call (e.g., `Vector2D_add(NULL, &v1, v2)`).

3. **Fallback to Primitives**:
   If the type is a primitive, or if the type is a struct but no matching method is found, the compiler falls back to primitive C arithmetic (and emits an error if applied to a struct without overloads).

## Examples

```nora
pub type Vector2D = struct {
    x: i32
    y: i32
}

pub fn (self: #Vector2D) add(other: Vector2D) Vector2D {
    return Vector2D{
        x: self.x + other.x,
        y: self.y + other.y,
    }
}

pub fn (self: #Vector2D) eq(other: #Vector2D) bool {
    return self.x == other.x && self.y == other.y
}

pub fn main() {
    var v1 = Vector2D{x: 10, y: 20}
    var v2 = Vector2D{x: 5, y: 15}

    // Resolves to: Vector2D_add(&v1, v2)
    var v3 = v1 + v2

    // Resolves to: Vector2D_eq(&v1, &v2)
    if v1 == v2 {
        panic("Equality failed")
    }
}

pub type CustomList = struct {
    data: List[i32]
}

// Custom indexing for reading
pub fn (c: #CustomList) index(i: i32) i32 {
    return c.data[i]
}

// Custom indexing for writing requires returning a pointer to the element
pub fn (c: &CustomList) index_mut(i: i32) &i32 {
    return &c.data[i]
}

pub fn index_example() {
    var list = alloc CustomList { data: make(List[i32], 0) }
    list.data.Push(10)
    
    // Resolves to: list.index(0)
    var val = list[0]
    
    // Resolves to: *list.index_mut(0) = 20
    list[0] = 20
}
```

## Edge Cases
- **Pointer Arithmetic**: Pointer arithmetic remains unsupported by default to uphold memory safety guarantees. If `+` is used on a pointer, it triggers a semantic error unless it is a leased struct type implementing the `add` method.
- **Order of Operations**: Operator overloading does not change standard operator precedence.
- **Short-Circuiting**: Comparison and logical operators preserve standard short-circuiting semantics.

## Future Considerations
- More complex custom desugaring for iterator loops via `in` operators.
