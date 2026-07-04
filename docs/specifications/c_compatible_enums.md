# Specification: C-Compatible Integer Enums (`enum(T)`)

**Status:** Proposed  
**Date:** 2026-07-04  
**Component:** Parser, Semantic Analyzer, Code Generator  

## Motivation
Currently, Nora reserves the `enum` keyword exclusively for **Sum Types** (tagged unions). When lowered to C11, these are emitted as composite structs (containing an `int _tag` and a `union _data`). 
While mathematically elegant, this makes it impossible to define native, FFI-compatible C enumerations (which must be raw 32-bit or 64-bit integers) without resorting to `struct { val: i32 }` wrappers and clunky constant functions (e.g., `sys.WGPUTextureFormat_BGRA8Unorm()`).

To maximize both **User Simplicity** and **Language Consistency**, we should avoid adding entirely new keywords or syntax (like `enum(i32)`). Instead, Nora's existing `enum` should be upgraded to intelligently handle C-compatible integers.

## Proposed Solution: Smart Enums (Auto-Elision)
We propose modifying the Semantic Analyzer and C11 Code Generator to auto-detect "payload-less" enums and seamlessly lower them to native C integers.

Additionally, the Parser must be updated to allow assigning explicit compile-time integer values (`= value`) to payload-less variants.

### Unified Syntax
Users will use the exact same `enum` syntax they already know:

```nora
// A C-compatible enum (auto-detected because it has no payloads)
pub type WGPUTextureFormat = enum {
    Undefined = 0x00000000,
    R8Unorm   = 0x00000001,
    BGRA8Unorm = 0x0000001B
}

// A standard Sum Type (auto-detected because of the payload)
pub type State = enum {
    Idle,
    Running(progress: i32)
}
```

### Usage & Access
Accessing variants of the enum uses standard dot-notation, drastically improving code readability compared to global variables or functions:
```nora
var format: WGPUTextureFormat = WGPUTextureFormat.BGRA8Unorm
```

## Semantic Rules
1. **Auto-Detection**: If the Semantic Analyzer determines that an `enum` contains **zero payloads** across all its variants, it flags the type as a `PrimitiveEnum`.
2. **Explicit Values**: Variants may optionally be assigned an integer value (e.g., `= 27`). If omitted, the value auto-increments from the previous variant (starting at `0`), matching standard C behavior.
3. **Payload Mixing**: You cannot assign an explicit `= value` to a variant that *also* has a payload.

## Code Generation (Lowering to C11)
During AST-to-HIR lowering, the compiler checks if the `enum` is a `PrimitiveEnum`.
If it is, the compiler entirely skips generating the `struct / union` wrapper. Instead, it emits a standard C `typedef` to `int32_t` (or `int64_t` if the assigned values exceed 32-bit bounds):

**Nora Source:**
```nora
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
