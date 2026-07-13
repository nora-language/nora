# Platform-Conditional Manifest Configuration

## Title & Overview

This specification defines support for **platform-conditional sections** inside `nora.yaml`. It allows developers to declare different `native:` linking configurations, `dependencies:`, and `cflags` per target OS — directly in the manifest — without needing to maintain separate manifest files per platform.

---

## Motivation

Nora's current conditional compilation story is split across two layers:

- **Source code**: file-suffix constraints (`_windows.nr`) and `[cfg("os=X")]` attributes handle platform-specific functions and types.
- **Manifest**: The `native:` block is **flat and global** — there is no way to express "link against `ws2_32` only on Windows and `pthread` only on Linux" without hardcoding everything for one platform.

This is a real gap for cross-platform projects that use FFI. A project that wraps OS networking APIs needs different headers, linker flags, and dynamic libraries per platform. Today, developers must manually maintain multiple `nora.yaml` files or use workarounds. This feature eliminates that entirely.

---

## Syntax

### Platform-Conditional `native:` Block

The `native:` section gains an optional sub-key per OS name. Each OS block contains the same fields as the current `NativeConfig`. Fields at the top-level `native:` apply to **all platforms** and are merged with the platform-specific block.

```yaml
name: my_project
version: 1.0.0
entry: src/main.nr

native:
  # Global — applied on all platforms
  cflags: ["-Wall"]

  # Platform-specific overrides and additions
  windows:
    headers: ["winsock2.h", "windows.h"]
    dynamic_libs: ["ws2_32", "kernel32", "advapi32"]
    cflags: ["-municode", "-D_WIN32_WINNT=0x0601"]

  linux:
    headers: ["sys/socket.h", "unistd.h", "pthread.h"]
    dynamic_libs: ["pthread", "dl", "m"]
    cflags: ["-lpthread"]

  darwin:
    headers: ["sys/socket.h"]
    dynamic_libs: ["m"]
    cflags: ["-framework", "CoreFoundation"]
```

### Platform-Conditional `dependencies:`

Individual dependencies can declare a `platform:` key to restrict them to a single OS. Dependencies without a `platform:` key are always resolved.

```yaml
dependencies:
  winapi_ffi:
    path: ../winapi_ffi
    version: 1.0.0
    platform: windows   # Only resolved and linked on Windows

  posix_utils:
    path: ../posix_utils
    version: 1.0.0
    platform: linux

  shared_lib:
    path: ../shared
    version: 1.0.0
    # No platform key — resolved on all targets
```

---

## Valid Platform Keys

| Key       | Matches `TargetOS` |
|-----------|--------------------|
| `windows` | `windows`          |
| `linux`   | `linux`            |
| `darwin`  | `darwin`           |
| `wasm`    | `wasm`             |
| `wasi`    | `wasi`             |

---

## Semantics

### Merging Strategy

Platform-specific `native:` blocks are **merged additively** with the global `native:` block. Lists are concatenated; scalar values (like `compiler`) are overridden by the platform-specific value.

| Field Type | Merge Behavior |
|---|---|
| Lists (`cflags`, `headers`, `dynamic_libs`, `static_libs`, `include_dirs`, `lib_dirs`, `source_files`) | Concatenated: `global + platform_specific` |
| Scalars (`compiler`, `opt_release`, `opt_debug`, etc.) | Platform-specific **overrides** global if present; otherwise global is used |

**Example merge result** (targeting `linux`):
```yaml
# Input
native:
  cflags: ["-Wall"]
  linux:
    cflags: ["-lpthread"]
    dynamic_libs: ["pthread"]

# Resolved for linux:
native:
  cflags: ["-Wall", "-lpthread"]
  dynamic_libs: ["pthread"]
```

### Dependency Resolution

Dependencies with a `platform:` key that does not match the current `TargetOS` are **completely excluded** from the dependency graph. They are not loaded, not compiled, and not linked. Their exported symbols are not available in the source code — any reference to them from unconditional source code will produce an `undefined identifier` error.

### Target OS Resolution

The same resolution order as conditional compilation:
1. `--target <platform>` CLI flag.
2. `runtime.GOOS` (host OS) if no `--target` flag is provided.

---

## Type Rules

- A platform-conditional dependency's exported symbols are only in scope on the matching platform. Cross-referencing them from unconditional code is the developer's responsibility — the compiler treats them as simply undefined on non-matching targets.

---

## Lease Rules

No change. Platform filtering at the manifest level happens before the Topological Lease Solver sees any code.

---

## Examples

### Example 1: Cross-Platform Socket Library Wrapper

```yaml
name: mysocket
version: 1.0.0
entry: src/lib.nr

native:
  windows:
    headers: ["winsock2.h"]
    dynamic_libs: ["ws2_32"]
    cflags: ["-D_WIN32_WINNT=0x0601"]
  linux:
    headers: ["sys/socket.h", "netinet/in.h"]
    dynamic_libs: ["pthread"]
  darwin:
    headers: ["sys/socket.h"]
```

### Example 2: Platform-Specific Dependency

```yaml
dependencies:
  win_registry:
    path: ../win_registry_ffi
    version: 1.0.0
    platform: windows   # Never compiled or linked on Linux/macOS
```

```nora
// src/main.nr — this would cause a compile error on Linux
import "win_registry"

pub fn main() i32 {
    win_registry.read_key("HKLM\\Software")
    return 0
}
```

To write this correctly, guard the call with `[cfg]`:
```nora
import "win_registry"

[cfg("os=windows")]
pub fn read_config() {
    win_registry.read_key("HKLM\\Software")
}
```

---

## Edge Cases

| Scenario | Behavior |
|---|---|
| Unknown platform key in `native:` (e.g., `native.myos:`) | Silently ignored |
| Platform-specific `compiler:` override | Overrides the global `compiler` only for that platform |
| Dependency `platform:` set to unknown OS | Dependency is never resolved on any target |
| Both global and platform lists have duplicates | Duplicates are preserved (same as manually listing them twice in `cflags`) |

---

## Errors & Diagnostics

| Error | Cause |
|---|---|
| `undefined identifier: <name>` | Code imports a symbol from a platform-excluded dependency |
| `package not found: <name>` | A dependency with a matching `platform:` key has an invalid `path` |
| `unknown native platform key: <key>` | A key under `native:` is not a known OS name and not a known field name |

---

## Future Considerations

- **Architecture-based conditions**: `native.linux_amd64:`, `native.linux_arm64:` for fine-grained arch targeting.
- **Multiple platform keys per block**: `platform: [windows, linux]` on a dependency to share a single block across multiple targets.
- **Exclusive vs. additive mode**: A `native.replace:` variant that fully replaces rather than merges the global config.
