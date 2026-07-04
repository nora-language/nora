# Investigation: C Header Collision with `extern fn`

**Status:** Investigated / Workaround Applied
**Date:** 2026-07-04
**Component:** `pkg/codegen` (C Backend Generator)

## Problem
When implementing the Win32 window wrapper in `nora_wgpu`, defining raw Win32 APIs (e.g., `pub extern fn PeekMessageA`) and structures (`pub type MSG`) inside a Nora `.nr` file caused catastrophic C compilation failures during `go run ../pkg/cmd/nora run --example triangle`. The C compiler threw numerous `conflicting types` and `typedef redefinition` errors.

## Reproduction
1. In any Nora file, define a Win32 struct and an extern function:
   ```nora
   pub type MSG = struct { ... }
   pub extern fn PeekMessageA(lpMsg: ptr, hWnd: ptr, filterMin: i32, filterMax: i32, removeMsg: i32) i32
   ```
2. Ensure the project links against a C header that implicitly includes `<windows.h>` (such as `wgpu-native`'s `webgpu.h`).
3. Compile the project. The C compiler will fail when compiling the generated `out_pkg_*.c` files because it encounters redefined types.

## Root Cause
Nora's C backend (`hir_codegen.go` and `generator.go`) currently emits C declarations for *all* `pub type` and `pub extern fn` definitions unconditionally into the shared `out.h` header file. 
For example, Nora generates:
```c
typedef struct MSG MSG;
int PeekMessageA(void* lpMsg, void* hWnd, int filterMin, int filterMax, int removeMsg);
```
When `webgpu.h` includes `<windows.h>`, the C compiler parses Microsoft's official definitions of `MSG` and `PeekMessageA`. Because Nora's generated prototypes map to `void*` and `int` rather than the exact `HWND` and `BOOL` macros used in `windows.h`, the C compiler detects a signature conflict and aborts the build.

## Fix / Workaround
**Immediate Workaround:** 
We bypassed the conflict by creating a separate native C wrapper (`src/sys/win32.c`) that includes `<windows.h>` and exposes clean, prefix-isolated functions (e.g., `Nora_CreateWindowEx`). By declaring `pub extern fn Nora_CreateWindowEx` in Nora, the generated C headers no longer conflict with any existing Windows symbols.

**Proposed Language Solutions (Future):**
1. **The `[NoEmit]` Attribute:** Add a compiler attribute for the AST and C-Generator that instructs Nora *not* to emit a C declaration for a specific `extern fn` or `type`, assuming it will be provided by an included C header.
   Example:
   ```nora
   [NoEmit]
   pub extern fn PeekMessageA(lpMsg: ptr, hWnd: ptr, filterMin: i32, filterMax: i32, removeMsg: i32) i32
   ```
2. **Dynamic Loading:** Implement robust `LoadLibrary` / `GetProcAddress` bindings in `std/ffi` to resolve OS-specific APIs at runtime, avoiding C headers entirely.

## Validation
By isolating the `windows.h` include inside `win32.c` and removing the conflicting Nora extern definitions, the `triangle` example compiled successfully and correctly executed the Win32 message loop without any C-level type collisions.
