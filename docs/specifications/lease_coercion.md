# Specification: Primitive Lease Coercion

## Title & Overview
Implicit coercion of primitive (copy-by-value) read-only or write leases to their base primitive types.

## Motivation
Previously, the Nora compiler failed to automatically coerce const borrows (like `#i32` or `#i64`) when assigning or returning values, forcing programmers to use a `+ 0` math hack to trigger a copy of the primitive value. Since primitive types (integers, floats, booleans) do not have move/drop semantics (they are copy-by-value), copying them from a lease is safe and ergonomic.

## Syntax
No syntax changes. Developers can pass or assign `#x` (where `x` is `i32`, `bool`, etc.) directly to variables or parameters expecting `i32` or `bool`.

## Semantics
When a value of lease type `#T` or `&T` is assigned or passed to a target expecting `T`, if `T` is a copy-by-value primitive type, the compiler implicitly dereferences the lease and copies the value.

## Type Rules
```text
Γ ⊢ expr : #T   (or &T)
T is a copy-by-value (non-owned) primitive type
------------------------------------------------
Γ ⊢ expr : T  (via implicit coercion)
```

## Lease Rules
Since `T` is a copy-by-value type, copying it does not consume or move the lease source. The source remains fully active.

## Examples
```nora
fn test_coercion(a: #i32) i32 {
    return a  // Coercion from #i32 to i32
}

fn main() {
    var val = 42
    var copied = test_coercion(#val)
}
```

## Edge Cases
- **Owned types**: Structs, sum types, and other owned types still require explicit moves (`@`) or conversions and do not undergo implicit coercion.
- **Generic types**: During generic type checking, generic parameters are treated as potentially owned. Coercion is resolved after specialization to concrete copy-by-value types.

## Errors & Diagnostics
No new diagnostics are introduced. If `T` is owned, the compiler continues to report standard "cannot move borrowed value" or "type mismatch" errors.

## Future Considerations
We might expand implicit coercion to other copyable structures if Nora introduces a copy trait or clone mechanism in the future.
