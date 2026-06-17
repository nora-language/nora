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
pub fn add_five(x: i32) i32 {
    return x + 5
}
```

### 2. Simple Custom Attributes

You can attach arbitrary simple identifiers as metadata for compiler plugins or reflection.

```nora
[simple]
type MyStruct = struct {
    x: i32
}
```

### 3. Parameterized Custom Attributes

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
