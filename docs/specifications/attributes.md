# Compiler Attributes

## Overview

Nora supports declarative metadata tags called Attributes (or Directives) attached to functions, structs, and variables. Attributes use the `[identifier]` syntax and instruct the compiler, optimizer, or code-generator to apply specific behaviors to the annotated symbol.

## Motivation

There are cases where the language syntax itself is not sufficient to dictate low-level optimizations (like inlining a function) or external metadata (like defining custom serialization rules or FFI bindings). Attributes provide an extensible, non-intrusive way to configure compiler features.

## Syntax

Attributes are declared inside square brackets `[...]` immediately preceding the target definition.

### 1. Built-in Attributes: `[inline]`

The `[inline]` attribute acts as a strong hint to the compiler's code-generator (and the underlying C compiler) to avoid function call overhead by substituting the function's body directly at the call site.

```nora
[inline]
pub fn fast_add(a: i32, b: i32) i32 {
    return a + b
}
```

### 2. Built-in Attributes: `[vector_size(N)]`

The `[vector_size(N)]` attribute enables hardware-native SIMD optimizations for struct types. It instructs the compiler to generate C `__attribute__((vector_size(N)))` extensions for the struct, aligning the data to `N` bytes and enabling native CPU SIMD instructions for arithmetic operators (`+`, `-`, `*`, `/`) and equality checks.

```nora
[vector_size(32)]
pub type Vec4d = struct {
    x: f64, y: f64, z: f64, w: f64
}
```

### 3. Built-in Attributes: `[NoEmit]`

The `[NoEmit]` attribute instructs the C-generator to skip emitting C declarations (forward declarations, struct bodies, and function prototypes) for specific structs and `extern fn`s. This is crucial for preventing type collisions when a C header (like `<windows.h>`) already defines these symbols.

```nora
[NoEmit]
pub type MSG = struct {
    hwnd: ptr,
    message: i32
}

[NoEmit]
pub extern fn PeekMessageA(lpMsg: ptr, hWnd: ptr, filterMin: i32, filterMax: i32, removeMsg: i32) i32
```

### 3. Built-in Attributes: `[intrinsic("name")]`

The `[intrinsic("name")]` attribute marks a function as a compiler intrinsic. Instead of generating a standard C function call, the compiler's code generator will intercept calls to this function and substitute them with hardcoded, optimized inline C expressions. This is used extensively in the `ffi` package (e.g., `borrow_to_raw` and `mut_borrow_to_raw`) to emit inline pointer address-of operators (`&`) for primitive borrows.

```nora
[intrinsic("borrow_to_raw")]
[NoEmit]
pub fn BorrowToRaw[T](val: #T) ptr {
    return val
}
```

### 4. Built-in Attributes: `[repr("type")]`

The `[repr("type")]` attribute is specifically used on `enum` declarations to force the compiler to lower the enum into a primitive C integer type (e.g., `i32`, `u8`) rather than a tagged union struct. This provides 100% C ABI compatibility for FFI and drastically simplifies C interoperability.
To use `[repr]`, the enum must not contain any data payloads in its variants.

```nora
[repr("i32")]
pub type WGPUTextureFormat = enum {
    Undefined = 0,
    R8Unorm = 1,
    R8Snorm = 2
}
```

### 5. Built-in Attributes: `[native("type")]`

The `[native("type")]` attribute is used on `struct` declarations to indicate that the struct maps directly to a native C type (such as a compiler intrinsic like `__m256d`). When a struct is marked as `[native]`:
1. **Move Semantics:** The semantic analyzer treats the struct as a primitive, copyable value (like an `i32` or `f64`), rather than an owned type. This prevents "use of moved value" errors when passing the struct by value.
2. **Code Generation:** The C generator skips emitting `typedef struct` forward declarations for the struct, avoiding redefinition errors if the type is already defined in a C header.
3. **Equality Operators:** The C generator uses `memcmp` for equality checks (`a == b`) instead of generating invalid field-by-field comparisons.

```nora
[native("__m256d")]
pub type Vec4d = struct {
    v0: f64
    v1: f64
    v2: f64
    v3: f64
}
```

### 6. Simple Custom Attributes

You can attach arbitrary simple identifiers as metadata for compiler plugins or reflection.

```nora
[simple]
type MyStruct = struct {
    x: i32
}
```

### 7. Parameterized Custom Attributes

Attributes can carry string arguments. This is incredibly useful for providing metadata like custom JSON field names, routing paths for HTTP handler plugins, or FFI mapping names.

```nora
[custom("arg1", "arg2")]
fn my_func() {
    io.PrintLn("hello from my_func")
}
```

## Semantics

1.  **Placement:** An attribute must immediately precede the type, function, or variable declaration.
2.  **Chaining:** Multiple attributes can be stacked above a single declaration (e.g., `[inline] \n [custom("fast")]`).
3.  **Validation:** The compiler parses all attributes and attaches them to the AST nodes. If an attribute relies on a specific compiler plugin, and the plugin is missing, it may be ignored or trigger a warning. Built-in attributes like `[inline]` are natively evaluated during the lowering to HIR (High-level Intermediate Representation).

## Examples

### FFI Name Overrides (Future)
Attributes are the foundation for features like explicitly mapping a C function name that clashes with Nora keywords:

```nora
// Using the C function 'do_work' as 'execute' in Nora
[extern_name("do_work")]
extern fn execute() i32
```

## Errors & Diagnostics

*   **Misplaced Attribute:** Placing an attribute without a trailing valid declaration (e.g., at the end of a file or block) will result in a parsing syntax error.
