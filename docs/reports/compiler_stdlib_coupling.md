# Report: Compiler and Standard Library Coupling in Nora

## Overview
This report investigates the current coupling between the Nora compiler (`pkg/`) and the standard library (`std/`). In a fully decoupled architecture, the standard library should be treated as just another package, using the same rules, syntax, and features available to user-defined packages. However, the current Nora implementation relies on hardcoded paths, injected macros, and explicit C runtime links that tightly fuse the compiler to the standard library.

Decoupling the compiler from the standard library is a necessary step for making Nora viable for bare-metal, embedded, and `#![no_std]` environments.

---

## 1. Magic Semantic Bypasses ("God Mode" for `std/`)

### The Issue
In `pkg/semantic/analyzer.go`, the Semantic Analyzer explicitly searches for the string `"std/"` or `"stdlib_"` in the absolute file path of the code it is compiling. 

```go
if strings.Contains(filename, "std/") || strings.Contains(filename, "stdlib_") {
    targetType = types.I32
} else {
    sa.AddError(idx.Pos(), "cannot assign to immutable string index")
    return
}
```

If it detects that the code belongs to the standard library, it bypasses the compiler's strict write-permission rules. For example, it allows modifying immutable string indexes natively.

### Why it is problematic
The `std/` directory is serving as an implicit unsafe block for the entire language. Because this "unsafe" behavior is tied directly to the folder name `"std/"`, no user can ever write their own high-performance memory abstractions outside of the standard library. 

### Path to Decoupling
* Remove path-based `"std/"` checks from the semantic analyzer.
* Introduce a robust `[unsafe]` attribute or an explicit `__intrinsic` keyword to demarcate unsafe operations.
* This allows any package to perform low-level memory operations, provided they explicitly use the unsafe syntax.

---

## 2. Mandatory Prelude Injection

### The Issue
In `pkg/lsp/handler.go` and `pkg/cmd/nora/main.go`, the compiler automatically searches for, parses, and injects `std/prelude.nr` into the Abstract Syntax Tree (AST) of every single compiled project.

```go
// main.go snippet
l := lexer.New(string(preludeInput), preludePath)
p := parser.New(l)
preludeFile := p.Parse(preludePath)
prog.Files = append(prog.Files, preludeFile)
```

### Why it is problematic
The compiler relies on `std/prelude.nr` for fundamental language syntax. For instance, `for i in 0..10` naturally translates `0..10` into a `Range` struct. Because this struct is defined in `std/prelude.nr`, deleting or renaming the `std/` folder will cause the compiler to crash with a `Range type not found in prelude` error. A developer cannot easily write a bare-metal kernel without the `std/` folder.

### Path to Decoupling
* Treat fundamental types like `Range` or `Iterator` as core compiler primitives, or allow the compiler to function without them if they aren't used.
* Implement a `#![no_std]` or `--no-stdlib` flag that disables the automatic injection of `std/prelude.nr`.

---

## 3. Hardcoded C11 Runtime Links

### The Issue
Nora compiles down to C11 code, but the generated C code is not fully independent. Inside `pkg/codegen/generator.go`, the compiler explicitly hardcodes `#include "nora_runtime.h"` at the top of every generated C file. It also injects C macros like `NR_COOPERATIVE_YIELD_CHECKPOINT()` into every generated function block.

### Why it is problematic
The compiler assumes that the `std/runtime/` folder—which contains the C implementations for Nora's stackless fibers, channels, and memory leak trackers (`nr_mem_report`)—will always be present and compiled alongside the user's code. This fuses the language's concurrency model and memory allocator directly to the compiler's code generator.

### Path to Decoupling
* Stop hardcoding `#include "nora_runtime.h"` in the generator.
* Make the cooperative yield checkpoints optional or tied to specific project configurations in `nora.yaml`.
* Allow developers to swap out the fiber runtime (e.g., to use POSIX threads directly) by decoupling the concurrency primitives from the code generator.
