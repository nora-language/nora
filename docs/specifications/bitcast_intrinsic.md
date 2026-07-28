# Specification: Bitcast Intrinsic and Bitwise Floating-Point Casts

## Title & Overview
**Bitcast Intrinsic & Float Bitwise Casting**

This specification defines the behavior of the `[intrinsic("bitcast")]` compiler attribute and the standard library wrapper functions for performing strict bitwise casting (type punning) between floating-point numbers and unsigned integers of the exact same bit width.

## Motivation
Systems programming, fast inverse square root implementations, and format conversion (such as `f32` to half-precision `f16`) often require accessing the raw IEEE-754 mantissa and exponent bits of a floating-point number. Standard value casts (e.g. `u32(1.0)`) perform semantic conversions, resulting in the value `1`. Prior to this feature, raw bit extraction required utilizing C FFI blocks. This feature aims to provide safe, zero-cost, and robust native support for bitwise type punning directly in the Nora language.

## Syntax
The feature introduces an internal compiler attribute `[intrinsic("bitcast")]`. 
Users interface with this feature exclusively through the `math` package in the standard library.

```nora
package math

[intrinsic("bitcast")]
pub fn Float32bits(f: f32) u32 {
	return u32(f)
}

[intrinsic("bitcast")]
pub fn Float64bits(f: f64) u64 {
	return u64(f)
}

[intrinsic("bitcast")]
pub fn Uint32bitsToFloat(u: u32) f32 {
	return f32(u)
}

[intrinsic("bitcast")]
pub fn Uint64bitsToFloat(u: u64) f64 {
	return f64(u)
}
```

## Semantics
The `bitcast` intrinsic completely overrides the AST execution of the function body. During the Code-Generation phase, the compiler intercepts the call and outputs a C99 compound literal with a `union`, forcing a bit-exact type pun.

For example, `Float32bits(1.0)` is compiled into:
```c
((union { float from; uint32_t to; }){ .from = 1.0f }).to
```
This guarantees strict-aliasing safety and produces zero-overhead assembly instructions (often compiling out completely into raw register operations in GCC/Clang).

## Type Rules
The compiler validates the types by strictly mapping the arguments and return values.
- `f32` ↔ `u32`
- `f64` ↔ `u64`

Any attempt to apply the `bitcast` intrinsic to a function with mismatched bit-widths (e.g. `f64` to `u32`) will result in a C compilation failure due to misaligned union padding. (Currently handled entirely internally by the standard library).

## Lease Rules
Since primitive numeric types (`f32`, `f64`, `u32`, `u64`) are purely `Copy` types without heap allocations or topological dependencies, lease tracking ignores these values. The intrinsic bypasses lease rules completely.

## Examples
```nora
import "math"

pub fn main() {
    var float_val = f32(1.0)
    var raw_bits = math.Float32bits(float_val) // raw_bits == 0x3f800000
    
    var restored = math.Uint32bitsToFloat(raw_bits)
    // restored == 1.0
}
```

## Edge Cases
- **NaN / Infinity**: The exact bit representation of `NaN`, `-Infinity`, and `+Infinity` are preserved identically to the platform's IEEE-754 architecture implementation.
- **Endianness**: The bitwise integer representation inherits the endianness of the target platform (almost universally little-endian on standard systems).

## Errors & Diagnostics
Because this feature is encapsulated entirely within standard library APIs (`math`), users cannot trigger semantic errors specific to this feature. Standard type-mismatch errors will apply if a user passes an incorrect type into the `math` functions.

## Future Considerations
- Expand the `bitcast` intrinsic to support arbitrary `struct` to `struct` transmutes, provided the structs are proven to have identical byte sizes at compile time (similar to Rust's `std::mem::transmute`).
- Introduce an `@bitcast[T](val)` keyword if direct, ad-hoc struct transmutation becomes common.
