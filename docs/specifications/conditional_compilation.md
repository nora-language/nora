# Conditional Compilation

## Overview

Nora provides first-class support for conditional compilation, allowing developers to target specific operating systems or configurations cleanly without duplicating code or creating linker conflicts. This is achieved through **two orthogonal mechanisms** that can be combined:

1. **File-Suffix Build Constraints** — exclude entire files at load time.
2. **Attribute-Based `[cfg]` Constraints** — exclude individual top-level declarations at symbol-collection time.

---

## Motivation

When building cross-platform software or interoperating with native C APIs via FFI (Foreign Function Interface), the underlying system types, APIs, and constraints vary wildly between operating systems (e.g., Windows vs. Linux). Nora requires a systematic way to include or exclude specific files, functions, and types at compile time based on the active target OS — preventing namespace collisions and link-time errors.

---

## Syntax

### 1. File-Suffix Constraints

Any source file whose name ends with `_<os>.nr` is strictly constrained to that specific operating system. The file is either **included or excluded wholesale** — it never enters the parsing phase on a non-matching target.

**Valid OS suffixes:**

| Suffix         | Target OS          |
|----------------|--------------------|
| `_windows.nr`  | `windows`          |
| `_linux.nr`    | `linux`            |
| `_darwin.nr`   | `darwin` (macOS)   |
| `_wasm.nr`     | `wasm`             |
| `_wasi.nr`     | `wasi`             |

**Example:**
```
platform_windows.nr   ← compiled only on Windows
platform_linux.nr     ← compiled only on Linux
platform.nr           ← compiled on all platforms
```

> **Note:** Any suffix that does not match the known OS list (e.g., `_myfeature.nr`) is treated as a normal filename — it is included unconditionally on all targets.

---

### 2. Attribute-Based `[cfg]` Constraints

For finer-grained control within a **single file**, Nora provides the `[cfg(...)]` attribute. It is applied to individual **top-level declarations** and filters them out during the `CollectSymbols` phase.

**Supported targets:** `pub fn`, `pub type` (structs, sum types, interfaces, aliases).

**Syntax:**

```nora
[cfg("os=<target>")]
pub fn my_function() { ... }
```

**Valid condition keys:**

| Key    | Description                        | Example               |
|--------|------------------------------------|-----------------------|
| `os=`  | Match the current target OS string | `[cfg("os=windows")]` |

**Examples:**

```nora
[cfg("os=windows")]
pub fn get_handle() ptr { ... }

[cfg("os=linux")]
pub fn get_handle() ptr { ... }
```

---

### Multiple Conditions (AND Semantics)

Multiple condition strings within a single `[cfg]` attribute are evaluated with **AND semantics** — all conditions must hold for the declaration to be included.

```nora
// Included only when BOTH conditions are true
[cfg("os=linux", "os=wasi")]  // ← currently never true (AND of two OS conditions)
pub fn my_fn() { ... }
```

> **Caution:** Multiple `os=` conditions in the same `[cfg]` use AND semantics, meaning they can never both be satisfied simultaneously. To target multiple platforms, use separate declarations each with their own `[cfg]`.

---

### Unknown Condition Keys

Condition keys other than `os=` are **silently ignored and treated as passing**. This is intentional permissive behavior to allow future feature flags to be added without breaking existing code.

```nora
[cfg("arch=amd64")]  // ← "arch=" is unrecognized today; declaration is always included
pub fn my_fn() { ... }
```

---

## Semantics

### Compilation Pipeline Interaction

Conditional compilation happens in **two distinct phases** before type-checking or lease solving:

```
Source Files
    │
    ├── [Phase 1: File Loading]
    │       isPlatformCompatible(filename, targetOS)
    │       Files with non-matching _<os>.nr suffix → EXCLUDED (never parsed)
    │
    ├── [Phase 2: Symbol Collection — CollectSymbols]
    │       isCfgCompatible(attributes, targetOS)
    │       Declarations with non-matching [cfg] → EXCLUDED from symbol table
    │
    └── [Phase 3: Semantic Analysis, Lease Solving, Codegen]
            Only sees declarations that survived both filters
```

### File Loading Phase
During file collection, the compiler inspects every filename. If a file contains a known OS suffix that does not match the current `TargetOS`, the file is **entirely skipped** — it is never lexed, parsed, or analyzed.

### Symbol Collection Phase
During `CollectSymbols`, the compiler inspects every top-level `FunctionStatement` and `TypeStatement`. If a `[cfg("os=X")]` attribute is present and `X` does not match the `TargetOS`, that declaration is **completely skipped** as if it were never written.

### Target OS Resolution
The target OS is resolved in the following priority:

1. `--target <platform>` CLI flag (e.g., `nora build --target windows-amd64`)
2. Host operating system (`runtime.GOOS`) — used as the default when no flag is given.

---

## Type Rules

- Declarations filtered by a non-matching `[cfg]` **do not exist in the symbol table** for that target.
- Referencing a `[cfg]`-excluded function from a non-`[cfg]`-guarded callsite results in a standard `undefined identifier` error.
- The error message is identical to referencing any other undefined symbol — there is **no special diagnostic** for "this exists on another platform."

---

## Lease Rules

Conditional compilation is fully resolved before the Topological Lease Solver phase. The solver only ever sees code that survived both filters. **No special lease rules apply.**

---

## Examples

### Example 1: Platform File Split (File-Suffix)

```nora
// main.nr — compiled on all platforms
package main

pub fn main() i32 {
    print_os()
    return 0
}

// platform_windows.nr — compiled only on Windows
package main

pub fn print_os() {
    io.PrintLn("Running on Windows")
}

// platform_linux.nr — compiled only on Linux
package main

pub fn print_os() {
    io.PrintLn("Running on Linux")
}
```

*Depending on the target, only one `print_os` function is compiled. The other file is not even loaded.*

---

### Example 2: In-File `[cfg]` Splitting

```nora
package main
import "std/io"

[cfg("os=windows")]
pub fn spawn_process() {
    // Windows-specific: CreateProcess via FFI
}

[cfg("os=linux")]
pub fn spawn_process() {
    // POSIX-specific: fork/exec via FFI
}
```

---

### Example 3: Negative Test (Expected Failure)

```nora
// linux_only_func_linux.nr
package mylib

pub fn linux_only_func() { ... }

// main.nr (compiled on Windows → linux_only_func_linux.nr is excluded)
package mylib

pub fn main() i32 {
    linux_only_func() // ERROR: undefined identifier 'linux_only_func'
    return 0
}
```

---

## Edge Cases

| Scenario | Behavior |
|---|---|
| File has `_linux.nr` suffix AND contains `[cfg("os=windows")]` on a function | File excluded entirely on non-Linux. On Linux, the file is included but the Windows-cfg'd function is excluded. The function is **unreachable on all platforms**. |
| File suffix does not match any known OS (e.g., `_myfeature.nr`) | Treated as a normal file — included unconditionally. |
| `[cfg]` applied to a non-top-level declaration (e.g., inside a function) | **Undefined behavior / currently unsupported.** Only top-level `fn` and `type` are filtered. |
| Multiple `[cfg]` attributes stacked on one declaration | Only the **first** `cfg` attribute returned by `GetAttribute` is evaluated. |
| `TargetOS` is empty string | Both file-suffix filtering and `[cfg]` filtering are **disabled** — all declarations are included. |

---

## Errors & Diagnostics

| Error | Cause |
|---|---|
| `undefined identifier: <name>` | A function or type excluded by `[cfg]` or file-suffix was referenced from always-compiled code. |
| `duplicate definition: <name>` | Two declarations with the same name both lack `[cfg]` guards on the same target. |

---

## Future Considerations

- **Logical operators**: Expanding `[cfg]` to support `not(...)`, `any(...)`, `all(...)` for composable conditions.
- **Architecture constraints**: `[cfg("arch=amd64")]`, `[cfg("arch=arm64")]`.
- **Feature flags**: `[cfg("feature=my_feature")]` for opt-in capability flags beyond OS targeting.
- **Struct field-level `[cfg]`**: Allowing platform-specific fields inside struct definitions.
- **Diagnostic improvements**: A dedicated "platform-excluded symbol" error with a hint pointing to the file/declaration that defines it on other targets.
