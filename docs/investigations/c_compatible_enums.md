# Specification: C-Compatible Integer Enums (`enum(T)`)

**Status:** Resolved  
**Date:** 2026-07-06  
**Component:** Parser, Semantic Analyzer, Code Generator  

## Motivation
Currently, Nora reserves the `enum` keyword exclusively for **Sum Types** (tagged unions). When lowered to C11, these are emitted as composite structs (containing an `int _tag` and a `union _data`). 
While mathematically elegant, this makes it impossible to define native, FFI-compatible C enumerations (which must be raw 32-bit or 64-bit integers) without resorting to `struct { val: i32 }` wrappers and clunky constant functions (e.g., `sys.WGPUTextureFormat_BGRA8Unorm()`).

To maximize both **User Simplicity** and **Language Consistency**, we should avoid adding entirely new keywords or syntax (like `enum(i32)`). Instead, Nora's existing `enum` should be upgraded to intelligently handle C-compatible integers.

## Implemented Solution: `[repr("type")]` Attributes
We modified the Semantic Analyzer and C11 Code Generator to support the `[repr("type")]` attribute on enums, lowering them to native C integers.

Additionally, the Parser was updated to allow assigning explicit compile-time integer expressions (`= value`) to payload-less variants.

### Unified Syntax
Users will use the exact same `enum` syntax they already know, but opt-in to primitive lowering via the `[repr]` attribute:

```nora
// A C-compatible enum explicitly backed by an i32
[repr("i32")]
pub type WGPUTextureFormat = enum {
    Undefined = 0,
    R8Unorm   = 1,
    BGRA8Unorm = 1 << 4 | 5
}
```

### Usage & Access
Accessing variants of the enum uses standard dot-notation, drastically improving code readability compared to global variables or functions:
```nora
var format: WGPUTextureFormat = WGPUTextureFormat.BGRA8Unorm
```

## Semantic Rules
1. **Validation**: If the Semantic Analyzer detects a `[repr("type")]` attribute, it verifies the type is a primitive integer (e.g., `i32`, `u8`).
2. **Payload Restriction**: You cannot assign a payload to any variant in a primitive enum.
3. **Explicit Values**: Variants may optionally be assigned an integer expression (e.g., `= 27`). If omitted, the value auto-increments from the previous variant (starting at `0`), matching standard C behavior. The expressions are fully evaluated at compile-time.

## Code Generation (Lowering to C11)
During AST-to-HIR lowering, the compiler checks if the `enum` is a `PrimitiveEnum`.
If it is, the compiler entirely skips generating the `struct / union` wrapper. Instead, it emits a standard C `typedef` to the specified backing type:

**Nora Source:**
```nora
[repr("i32")]
pub type WGPUTextureFormat = enum {
    Undefined = 0,
    BGRA8Unorm = 27
}
```

**Emitted C11 Code:**
```c
typedef int32_t WGPUTextureFormat;
#define WGPUTextureFormat_Undefined 0
#define WGPUTextureFormat_BGRA8Unorm 27
```

## Benefits
- **Zero Syntax Churn**: Users do not have to learn a new `enum(T)` syntax. The language remains highly unified.
- **Zero-Cost FFI**: 100% ABI compatibility with C libraries like WebGPU.
- **Superior Readability**: Eliminates the need for generated `_` prefixed global functions. `WGPUTextureFormat.BGRA8Unorm` is extremely readable.
