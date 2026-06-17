# Specification: `#![no_std]` (No Standard Library) Environments

## Title & Overview
The `#![no_std]` Equivalent (No Standard Library) environments feature provides the ability to bypass the automatic injection of the Nora Standard Library (`std`) and its `prelude.nr`. This allows developers to write bare-metal code, operating system kernels, bootloaders, and resource-constrained embedded software natively in Nora.

## Motivation
By default, the Nora compiler automatically injects `std/prelude.nr` into the root scope of all projects. The prelude automatically includes definitions for basic primitives (like `Range`), dynamic collections (like `Vector`), and concurrency primitives. For systems programming where heap allocations, standard I/O, or a fiber runtime might not exist, automatic inclusion of these features creates unacceptable bloat or compilation failures.

This specification defines how to tell the Nora compiler and Language Server to strictly omit the standard library, returning total control over all imported scopes to the developer.

## Syntax & Configuration
Nora does not use in-file attributes (like Rust's `#![no_std]`) to disable the standard library. Instead, this is considered a global project configuration and is declared at the build or project manifest level.

### 1. `nora.yaml` Manifest (Preferred)
A `no_stdlib` boolean key can be specified in the project's `nora.yaml` configuration. When set to `true`, the compiler and the Language Server (LSP) will not inject the prelude.

```yaml
name: my_os_kernel
version: 1.0.0
language: 0.1.0
entry: src/main.nr
output: kernel.bin
no_stdlib: true
```

### 2. CLI Overrides
The compiler's CLI interface exposes a `--no-stdlib` flag that allows forcing this behavior, even if `no_stdlib` is not defined in `nora.yaml`.

```bash
Nora build --no-stdlib src/main.nr
Nora run --no-stdlib src/main.nr
Nora test --no-stdlib
```

## Semantics

1. **Empty Global Scope**: When `no_stdlib` is activated, the global scope is completely clean. Only explicitly imported modules are available.
2. **Missing Built-ins**: Structs like `String`, `Vector`, and `Result` are completely absent. The developer must re-implement them or import a separate minimal `core` library if they need them.
3. **LSP Accuracy**: The Nora Language Server Protocol natively reads the `nora.yaml` `no_stdlib` flag. If it is `true`, auto-completion, hover-documentation, and go-to-definition will not surface standard library prelude items, ensuring accurate intelligent code assistance for bare-metal developers.

## Type Rules & Built-in Syntax Mapping
Certain syntax sugars in Nora map to underlying library structs.

### Range Expressions
When compiling a range expression such as `0..10`, the Nora compiler looks for a struct named `Range` in the local/global scope.
* **Standard Behavior**: The compiler automatically uses `std.Range` provided by the prelude.
* **No-Std Behavior**: Since the prelude is not injected, the compiler will fail with: `compiler error: Range type not found in scope (required for range syntax)`.

To use the `..` operator in a `no_stdlib` environment, the developer must explicitly define or import a `Range` struct.

```nora
// User-provided definition to satisfy the compiler
pub type Range = struct {
    Start: i32
    End:   i32
}

pub fn main() {
    let r = 0..10 // Now perfectly valid without the standard library
}
```

## Lease Rules
`no_stdlib` does not alter Nora's Topological Lease Solver (`pkg/topology`). All ownership (`@`), read-only borrows (`#`), and mutable borrows (`&`) continue to enforce compile-time memory safety without exceptions. RAII (`drop()`) will still be inserted where necessary.

## Examples

### A Bare-Metal Kernel Entry Point
```yaml
# nora.yaml
name: bootloader
version: 0.1.0
entry: src/boot.nr
no_stdlib: true
```

```nora
// src/boot.nr
// No standard library, no hidden allocations.

[export]
pub fn _start() {
    // Platform-specific assembly or raw memory writes
    let vga_buffer: &u16 = [unsafe] cast[&u16](0xB8000)
    vga_buffer.* = 0x0F41 // Print 'A' in white text
    
    loop {
        // Halt loop
    }
}
```

## Edge Cases
1. **Accidental Stdlib Import**: A developer in a `no_stdlib` environment can technically still write `import "std/fmt"`. The compiler will successfully import it if the `std` library is in the compiler's search path. `no_stdlib` prevents *automatic injection*, but does not forcibly sandbox the `std/` directory.

## Errors & Diagnostics
* `Range type not found in scope (required for range syntax)`
  * Emitted when attempting to use the `..` operator without a `Range` struct in scope.
* Unrecognized type errors for `Vector`, `String`, etc., when attempting to use standard collections without the prelude.

## Future Considerations
* **Core vs Std**: A future split of the standard library may formalize a `core` library (which only contains data structures and traits that do not require an OS or an allocator) and a `std` library (which contains OS dependencies). `no_stdlib` currently removes both.
