# Investigation: Global Const Support

**Status:** Completed (Implemented 2026-07-06)  
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
While Nora's parser *did* successfully parse global `pub var` declarations, relying on `var` means these values can theoretically be mutated at runtime, which is highly unsafe for FFI constants. Furthermore, the semantic analyzer previously failed to resolve these global variables across package boundaries anyway.

## The Workaround (Depreciated)
To unblock `nora_wgpu`, we previously forced the binding generator (`generate_bindings.py`) to emit **constant functions** instead of global variables:
```nora
pub fn WGPUTextureFormat_BGRA8Unorm() WGPUTextureFormat {
    return WGPUTextureFormat { val: 27 }
}
```

## Resolution: Native `const` Implemented
The compiler has been updated to officially support the `const` keyword natively.

1. **Global & Local `const` Declarations:**
   The `const` keyword is now supported at both the package level and local function scope. 
   ```nora
   pub const WGPUTextureFormat_BGRA8Unorm = WGPUTextureFormat { val: 27 }
   ```
2. **Immutability Enforcement:**
   The semantic analyzer automatically assigns `WritePerm = false` for all `SymConst` instances. Any attempts to assign a new value to a constant at compile time will throw a semantic violation error (`cannot assign to X (it is a Constant)`).
3. **Cross-Package Constants:**
   `pub const` declarations are automatically exposed via the module's SymbolScope and can be safely imported and used by dependent packages seamlessly without function-call overhead. 
4. **Zero-Overhead C Emission:**
   Constants are natively emitted into the target C source code as standard variable definitions that avoid the dynamic C-initialization compilation errors that C compilers traditionally complain about when assigning non-literal values to `const` types at the file level.
