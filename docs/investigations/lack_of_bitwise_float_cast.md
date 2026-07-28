# Investigation: Lack of Bitwise Float Casting (`bitcast` / `transmute`)

## Status
Completed

## Problem
During the implementation of **Phase 36: Packing & Quantization** for `nora_math`, we needed to convert `f32` values to `f16` (half-floats). This requires extracting the raw IEEE-754 exponent and mantissa bits from an `f32`.

However, the Nora standard library currently lacks built-in functions for bitwise floating-point casting (such as `math.Float32bits(f)` found in Go, or `std::mem::transmute` in Rust, or `bitcast` keywords). Furthermore, because `nora_math` operates with `allow_unsafe: false` in `nora.yaml`, we cannot resort to unsafe raw pointer punning (e.g., `*(*u32)(&f)`) even if the syntax supported it natively.

Standard type casting in Nora (e.g., `u32(my_float)`) performs a semantic value conversion (e.g., truncating `1.0` to `1`) rather than a bitwise reinterpretation, resulting in the inability to access floating-point bits natively in pure safe Nora code.

## Reproduction
Attempting to cast a float to a uint32 directly loses the bit pattern:
```nora
var f = f32(1.0)
var u = u32(f) // Evaluates to 1, instead of 0x3f800000 (1065353216)
```
Attempting to call standard library conversions fails:
```nora
import "math"
var bits = math.Float32bits(f) // Error: package 'math' has no member 'Float32bits'
```

## Root Cause
Nora is a strictly typed language aiming for compile-time safety and memory safety. The AST-to-HIR lowering (`pkg/hir`) implements explicit `Cast` instructions, but these translate directly to C-style scalar conversions `(uint32_t)val` during the C11 Code-generation (`pkg/codegen`), which converts values, not bit patterns. 

A dedicated `bitcast` or `transmute` intrinsic has not yet been introduced into the compiler frontend, nor has a standard library wrapper been provided in `std/math`.

## Fix / Workaround Used
To bypass this language limitation without modifying the Nora compiler, we utilized Nora's FFI system to link a C-native implementation. 
We created a small C file (`native_pack.c`) that utilizes a C-style `union` for safe type-punning, which strictly adheres to C11 standards and preserves the exact bit patterns:
```c
#include <stdint.h>

uint32_t FloatBitsToUint(float f) {
    union { float f; uint32_t u; } u_val;
    u_val.f = f;
    return u_val.u;
}
```

We linked this file in `nora.yaml` under `native.source_files` and bound it to Nora using `extern fn`:
```nora
extern fn FloatBitsToUint(f: f32) u32
```

## Solution
This investigation has been resolved. We successfully implemented compiler-native bitwise floating-point casts, bypassing the need for C FFI bindings.

We introduced a new `[intrinsic("bitcast")]` attribute. 
The standard library (`core/math/bits.nr`) now provides native `Float32bits`, `Float64bits`, `Uint32bitsToFloat`, and `Uint64bitsToFloat` functions. During the code-generation phase (`pkg/codegen/expressions.go` and `pkg/codegen/hir_codegen.go`), the compiler intercepts these intrinsics and automatically emits safe C99 compound `union` literals directly into the generated binary, resulting in zero-overhead type punning.

This makes `native_pack.c` completely obsolete.
