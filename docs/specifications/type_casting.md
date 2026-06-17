# Type Casting

## Overview

Nora does not support implicit type coercion between numeric types, prioritizing safety and predictability over convenience. All type conversions must be explicit using the `type(value)` syntax.

## Syntax

Type casting uses a function-call-like syntax, where the target type is the function name and the value to cast is the argument.

```nora
var a_i32: i32 = 42
var a_i64 = i64(a_i32)
```

## Supported Casts

Nora supports casting between primitive numeric types and booleans.

### 1. Widening Casts
Casting to a type with a larger bit-width is always safe and preserves the exact value.

```nora
var a_u16: u16 = 100
var a_u32 = u32(a_u16)
```

### 2. Narrowing Casts
Casting to a type with a smaller bit-width truncates the higher-order bits.

```nora
var b_i64: i64 = 999999
var b_i32 = i32(b_i64)
```

### 3. Signed ↔ Unsigned Casts
Casting between signed and unsigned types of the same or different widths preserves the underlying bit pattern, but interpretation changes, which may lead to wrapping/overflow.

```nora
var c_i32: i32 = -5
var c_u32 = u32(c_i32) // Underflows to a large positive number
```

### 4. Float ↔ Integer Casts
Casting a floating-point number to an integer truncates the fractional part (rounds toward zero). Casting an integer to a float may lose precision if the integer cannot be represented exactly in the floating-point format.

```nora
var d_f64: f64 = 3.85
var d_i32 = i32(d_f64) // Results in 3
```

### 5. Boolean ↔ Integer Casts
Booleans can be cast to and from integers. `true` casts to `1`, and `false` casts to `0`. When casting an integer to a boolean, `0` becomes `false` and any non-zero value becomes `true`.

```nora
var e_bool = true
var e_i32 = i32(e_bool) // 1

var e_val: i32 = 0
var e_flag = bool(e_val) // false
```

## Errors & Diagnostics

*   **Invalid Cast:** Attempting to cast between fundamentally incompatible types (e.g., casting a `struct` to an `i32` without a defined conversion operator) yields a compile-time error.
