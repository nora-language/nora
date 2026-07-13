# Conditional Compilation

## Overview
Nora provides first-class support for conditional compilation, allowing developers to target specific operating systems or configurations cleanly without duplicating code or creating linker conflicts. This is achieved through two mechanisms: File-Suffix Build Constraints and Attribute-Based Conditional Compilation (`[cfg]`).

## Motivation
When building cross-platform software or interoperating with native C APIs via FFI (Foreign Function Interface), the underlying system types, APIs, and constraints vary wildly between operating systems (e.g., Windows vs. Linux). Nora requires a systematic way to include or exclude specific files, functions, and structs at compile time based on the active target OS to prevent namespace collisions and link-time errors.

## Syntax

### 1. File-Suffix Constraints
Any file ending in `_<os>.nr` is strictly constrained to that specific operating system.

Valid OS suffixes include:
- `_windows.nr`
- `_linux.nr`
- `_darwin.nr`
- `_wasm.nr`
- `_wasi.nr`

### 2. Attribute-Based `[cfg]` Constraints
For finer-grained control within a single file, Nora provides the `[cfg(...)]` attribute syntax. It can be applied to top-level `FunctionStatement` and `TypeStatement` declarations. 

```nora
[cfg("os=windows")]
pub fn get_os_name() str {
    return "Windows"
}
```

*Note: The attribute currently requires the condition to be a string literal.*

## Semantics

- **File Loading Phase:** During the file collection phase, the compiler inspects every file in the package. If a file contains a known OS suffix that does not match the current `TargetOS`, the file is entirely ignored. It does not enter the parsing phase.
- **Semantic Analysis Phase:** During the first pass of symbol collection (`CollectSymbols`), the compiler checks the attributes of every top-level node. If a `[cfg("os=X")]` attribute is present and `X` does not match the `TargetOS`, the node is completely skipped as if it didn't exist.
- **Target OS Resolution:** The target OS is inferred from the `Nora build --target <t>` flag. If omitted, it defaults to the host operating system (`runtime.GOOS`).

## Type Rules
- Function and Type signatures wrapped in non-matching `[cfg]` blocks do not exist in the symbol table. Attempting to reference them from another file will result in a standard `undefined identifier` error, just as if the code were never written.

## Lease Rules
- Conditional Compilation happens entirely before the Topological Lease Solver phase. The solver only sees the code that successfully survived the conditional filters. No special lease rules apply.

## Examples

**Example 1: Platform-Specific Implementations**
```nora
// In main.nr
pub fn main() i32 {
    print_os()
    return 0
}

// In platform_windows.nr
pub fn print_os() {
    print("Running on Windows")
}

// In platform_linux.nr
pub fn print_os() {
    print("Running on Linux")
}
```
*Depending on the target, only one `print_os` function is passed to the semantic analyzer, preventing duplicate definitions.*

**Example 2: In-File Configuration**
```nora
import "std/libc"

[cfg("os=windows")]
pub fn spawn_process() {
    // Windows specific logic
}

[cfg("os=linux")]
pub fn spawn_process() {
    // POSIX specific logic
}
```

## Edge Cases
- **Overlapping Configurations:** Using both `_windows.nr` on a file and `[cfg("os=linux")]` on a function inside it. The file will be skipped entirely on Linux, meaning the Linux function is never reached. On Windows, the file is parsed, but the function is skipped because it requires Linux. This renders the function completely unreachable on all platforms.
- **Unknown Suffixes:** Suffixes that do not match the known OS list (e.g., `_myfeature.nr`) are ignored and treated as normal files.

## Errors & Diagnostics
- **Undefined Identifier:** Calling a function excluded by the target OS results in `Error: undefined identifier: ...`
- **Duplicate Definition:** Defining two functions with the same name without `[cfg]` blocks results in a duplicate definition error.

## Future Considerations
- Expanding `[cfg]` to support logical operators like `not`, `any`, and `all`.
- Supporting architecture-based constraints (e.g., `[cfg("arch=amd64")]`).
- Extending `[cfg]` attributes to fields within a struct or individual variable declarations.
