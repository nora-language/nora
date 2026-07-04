# Investigation: Global Const Support

**Status:** Open  
**Date:** 2026-07-04  

## Context
While building the `nora_wgpu` bindings, we encountered a significant limitation when mapping WebGPU's native C `enum` values (e.g., `WGPUTextureFormat_BGRA8Unorm = 0x0000001B`) into the Nora language.

Currently, we are wrapping these C-enums as struct wrappers (e.g., `pub type WGPUTextureFormat = struct { val: i32 }`). 

## The Problem

### 1. Lack of Global `const` Immutability
To prevent users from writing error-prone magic numbers like `sys.WGPUTextureFormat { val: 27 }`, we need to expose the constants.
In Go or Rust, we would expose these as top-level immutable constants:
```nora
pub const WGPUTextureFormat_BGRA8Unorm: WGPUTextureFormat = WGPUTextureFormat { val: 27 }
```
While Nora's parser *does* successfully parse global `pub var` declarations (e.g. `pub var MyConst = ...`), relying on `var` means these values can theoretically be mutated at runtime, which is highly unsafe for FFI constants. Furthermore, as documented in `cross_package_var_bug.md`, the semantic analyzer currently fails to resolve these global variables across package boundaries anyway.

## The Workaround
To unblock `nora_wgpu`, we are forcing the binding generator (`generate_bindings.py`) to emit **constant functions** instead of global variables:
```nora
pub fn WGPUTextureFormat_BGRA8Unorm() WGPUTextureFormat {
    return WGPUTextureFormat { val: 27 }
}
```
While this is completely type-safe and eliminates magic numbers, it is syntactically verbose and slightly bloated (requiring function calls `()` instead of pure property access).

## Recommended Language Features
To make Nora a true systems programming language capable of replacing C/C++, the compiler team should investigate implementing the following:

1. **Global `const` Declarations:**
   Introduce the `const` keyword to allow static, compile-time instantiation of primitives and structs at the package level.
