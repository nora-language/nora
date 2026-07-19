# Investigation: WGSL Multiline String Parsing Bug

**Status**: Completed

## Problem
During the implementation of Phase 29 (Lighting Materials - Phong & PBR), embedding WGSL shader code inside a multiline string literal (`"..."`) caused the Nora compiler to fail with a large cascade of semantic errors. The compiler emitted errors such as `undefined identifier: group` and `undefined identifier: location`, which indicated that the compiler was attempting to parse the WGSL code as raw Nora syntax rather than treating it as a string literal.

## Reproduction
To reproduce the bug, assign a multiline string (without explicit line-continuation or escaping) to a variable or return it from a function:

```nora
pub fn PhongWGSL() str {
    return "
struct CameraUniforms {
    view_proj: mat4x4<f32>,
    model: mat4x4<f32>,
};
@group(0) @binding(0) var<uniform> camera: CameraUniforms;
"
}
```

Running `nora build` or `nora run` on a file containing this construct results in semantic errors like:
```
Error: undefined identifier: group
Error: undefined identifier: binding
Error: undefined identifier: location
```

## Root Cause
The Nora lexer/parser does not support unescaped multiline string literals using double quotes (`"`). When a newline is encountered before the closing quote, the string is implicitly terminated or the lexer loses synchronization. Consequently, the subsequent lines containing WGSL shader code (like `@group(0)`) are parsed as standard Nora source code. Because Nora doesn't recognize WGSL keywords or syntax, it triggers immediate semantic/parsing errors.

## Fix
To resolve the issue in the short term, the multiline string was collapsed into a single, continuous string literal on one line. For example:

```nora
pub fn PhongWGSL() str {
    return "struct CameraUniforms { view_proj: mat4x4<f32>, model: mat4x4<f32>, }; @group(0) @binding(0) var<uniform> camera: CameraUniforms;"
}
```

*Future Consideration*: The Nora compiler's `pkg/lexer` should be updated to properly support multiline strings (e.g., using backticks `` ` `` as in Go, or triple quotes `"""` as in Python/Rust) to allow for readable inline WGSL and large text blocks.

## Validation
By converting both `PhongWGSL()` and `PBRWGSL()` into single-line strings, the semantic errors vanished. The compiler successfully parsed, type-checked, and compiled the shader code into the executable, and both `examples/phong/main.nr` and `examples/pbr/main.nr` ran correctly without any errors.
