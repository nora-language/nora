# Investigation: Codegen Silently Drops `extern fn` Call When Return Type is an Enum and Value is Discarded

**Status:** Open

## Problem

Two distinct bugs were found in the Nora compiler, both triggered by the C-compatible enum refactor (`repr("i32")`/`repr("i64")`). Together, they caused `nora_wgpu`'s WebGPU triangle example to crash after exactly 3 rendered frames with a `NULL` texture handle.

---

## Bug 1 — `pkg/codegen`: Discarded `extern fn` call with enum return type becomes a no-op

### Description

When an `extern fn` is declared with a **`repr("i32")` or `repr("i64")` enum return type** and the caller **discards the return value**, the Nora codegen emits a **C type cast of the first argument** instead of an actual function call. The extern function is never invoked, silently becoming a no-op.

This bug was introduced during the C-compatible enum refactor, which changed many WebGPU FFI wrapper functions from returning struct wrapper types (e.g., `WGPUStatus { val: i32 }`) to returning `repr("i32")` enum types (e.g., `WGPUStatus`). Before the refactor, discarding the return value of a struct-returning `extern fn` worked correctly.

### Reproduction

The bug was discovered in `nora_wgpu` when `surface.Present()` stopped actually calling `wgpuSurfacePresent`, causing the WebGPU swapchain to exhaust all 3 triple-buffer images after exactly 3 frames (acquired but never presented), making the 4th call to `wgpuSurfaceGetCurrentTexture` return a `NULL` texture handle.

**Minimal reproduction in Nora:**
```nora
[repr("i32")]
pub type MyStatus = enum {
    Ok = 0,
    Err = 1
}

pub extern fn some_c_fn(handle: ptr) MyStatus

pub fn call_and_discard(handle: ptr) {
    some_c_fn(handle)  // return value discarded
}
```

**Generated C (buggy):**
```c
void call_and_discard(void* _env_ptr, void* handle) {
    sys_MyStatus _hir_tmp_0;
    _hir_tmp_0 = ((sys_MyStatus)handle);  // BUG: casts handle, never calls some_c_fn!
}
```

**Expected generated C:**
```c
void call_and_discard(void* _env_ptr, void* handle) {
    some_c_fn(handle);
}
```

### Root Cause

In `pkg/codegen/hir_codegen.go`, when a call to an `extern fn` has its return value discarded, the HIR lowers it into a `hir.Store` where the source (`Val`) is resolved as a `hir.Cast` expression. For a `repr("i32")` primitive enum return type, the codegen identifies it as a primitive-like value and routes through the enum cast path — `(sys_MyStatus)arg` — instead of emitting the actual function call `some_c_fn(arg)`.

### Workaround Applied

In `nora_wgpu/src/sys/wgpu.nr`, the declaration of `wgpuSurfacePresent` was changed to return `void` since the return value is not needed:

```nora
// Before (triggers the bug):
pub extern fn wgpuSurfacePresent(surface: ptr) WGPUStatus

// After (workaround):
pub extern fn wgpuSurfacePresent(surface: ptr) void
```

### Impact

Any `extern fn` declared with a `repr("i32")` or `repr("i64")` enum return type whose return value is **discarded at the call site** will silently become a no-op. Other `extern fn` bindings in `nora_wgpu` should be audited for this issue. All bindings that return an enum and whose return value is ignored should temporarily be changed to return `void` until the codegen is fixed.

### Required Fix

The codegen in `pkg/codegen/hir_codegen.go` must be fixed to correctly handle `hir.Store` where the source operand is a discarded call return value from an `extern fn` returning a `repr("i32")`/`repr("i64")` enum. A regression test should be added to `pkg/cmd/test/` verifying that an `extern fn` returning an enum is actually called when its return value is discarded.

---

## Bug 2 — `pkg/codegen`: Direct equality comparison between two `repr` enum values may not work

### Description

After the enum refactor, comparing two `repr("i32")` enum values with `!=` (e.g., `surf_tex.status != sys.WGPUSurfaceGetCurrentTextureStatus.SuccessOptimal`) was suspected to produce incorrect codegen or always evaluate to false/true, causing the status check to misbehave. This was observed when the status-check guard in `main.nr` was attempted using direct enum comparison and had to be removed because it introduced further `cast` undefined errors.

```nora
// Attempted — failed to compile/work correctly:
if surf_tex.status != sys.WGPUSurfaceGetCurrentTextureStatus.SuccessOptimal {
    continue
}
```

### Status

Not fully isolated. Superseded by Bug 1 as the primary crash cause. Needs a dedicated reproduction test in `pkg/cmd/test/c_compatible_enum_casts/` to verify that `==` and `!=` comparisons between two `repr` enum values produce correct C code and semantics.

---

## Validation

Bug 1 workaround was validated by running the `nora_wgpu` triangle example. Previously the example crashed after exactly 3 frames with a `wgpu-native` panic at `invalid texture`. After changing `wgpuSurfacePresent` to return `void`, the example ran continuously for 8+ minutes without any crash.
