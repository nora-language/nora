# Investigation: GPU Handle NULL-Release Panic via RAII Pre-Drop

**Status:** Resolved  
**Date:** 2026-07-09  
**Affected Component:** `nora_wgpu` library (`src/core/pipeline.nr`)  
**Classification:** Library convention bug (NOT a compiler codegen bug)

---

## Problem

When running the `nora_wgpu` cube example, the program panicked immediately after `MaterialBuilder.Build()` with a non-unwinding abort from inside `wgpu-native`:

```
thread 'unnamed' panicked at src\lib.rs:2994:5:
invalid pipeline layout
thread caused non-unwinding panic. aborting.
```

The error message "invalid pipeline layout" appeared to point at `CreateRenderPipeline`, creating the false assumption that the `PipelineLayout` handle was corrupt when passed to the render pipeline descriptor.

---

## Reproduction

```nora
// In gfx/pipeline_builder.nr Build():
var pipeline_layout = alloc core.PipelineLayout { handle: none }  // initial null handle
var layout_handle: ptr = none
if self.bind_group_layouts.Len[ptr]() > 0 {
    var pipeline_layout_desc = alloc sys.PipelineLayoutDescriptor { ... }
    pipeline_layout = self.device.CreatePipelineLayout(#pipeline_layout_desc)  // reassignment
    layout_handle = pipeline_layout.handle
}
```

---

## Root Cause

### Compiler Behavior (Correct)

The Nora topology solver emits a "pre-drop then reassign" pattern for any variable
reassignment where the variable already holds a live value. For the line:

```nora
pipeline_layout = self.device.CreatePipelineLayout(#pipeline_layout_desc)
```

The generated C is:

```c
{ core_PipelineLayout* _old = pipeline_layout;
  pipeline_layout = core_Device_CreatePipelineLayout(...);
  if (nr_df_pipeline_layout) {
    core_PipelineLayout_drop(NULL, _old);  // drops the INITIAL NULL-handle object
    nr_free(_old);
  }
}
```

This is **correct RAII behavior**: any previously-owned value must be released before
the variable is overwritten. The compiler cannot skip this drop because, in general,
the old value may hold real resources.

### Library Bug (Root Cause)

The initial value `alloc core.PipelineLayout { handle: none }` creates a `PipelineLayout`
object whose `handle` field is `NULL (0)`. When the pre-drop fires,
`core_PipelineLayout_drop` is called on this object:

```nora
// Before fix — in src/core/pipeline.nr:
pub fn (self: &PipelineLayout) drop() {
    sys.wgpuPipelineLayoutRelease(self.handle)  // Called with self.handle = NULL!
}
```

`wgpuPipelineLayoutRelease(NULL)` is **not null-safe** in wgpu-native. It triggers an
internal Rust `panic!()` at `src\lib.rs:2994` with the message `"invalid pipeline layout"`,
which propagates as a non-unwinding abort, crashing the entire process.

The misleading error message referred to the `NULL` handle passed to
`wgpuPipelineLayoutRelease`, NOT to the `RenderPipelineDescriptor.layout` field.

### Why ShaderModule Was Not Affected

`ShaderModule::drop()` already had a null-guard (added during an earlier investigation):

```nora
pub fn (self: &ShaderModule) drop() {
    if self.handle != none {   // null guard was already present
        sys.wgpuShaderModuleRelease(self.handle)
    }
}
```

The null guard was never applied to `PipelineLayout`, `BindGroupLayout`, `BindGroup`,
`ComputePipeline`, or `RenderPipeline`.

---

## Debugging Process

The panic stack trace pointed at `src\lib.rs:2994: invalid pipeline layout`, leading to
investigation of (all false trails):
1. BindGroupLayout handle validity — handle was valid (`0x6B2838B0`)
2. Struct layout size mismatches — all sizes matched (120 bytes for BGLE, 48 for PLD, etc.)
3. PipelineLayout lifetime vs CreateRenderPipeline call order — correct order confirmed

**Breakthrough:** Patched the generated `out_pkg_gfx.c` to add `printf` calls before and
after `CreatePipelineLayout`. Output showed:

```
[DBG] BGL array data ptr: 000002336ABDF770
[DBG] BGL handle[0]: 000002336B2838B0
[DBG] BGL count: 1
<CRASH HERE — printf after CreatePipelineLayout never printed>
```

This proved the crash happened **inside** `CreatePipelineLayout`, not `CreateRenderPipeline`.
The full stack trace from the debug binary confirmed:

```
18:  wgpuPipelineLayoutRelease     <-- crashing on the NULL pre-drop
23:  core_PipelineLayout_drop
24:  gfx_PipelineBuilder_Build     at pipeline_builder.nr:170
```

---

## Fix

Added null-guards to all GPU resource `drop()` functions in `src/core/pipeline.nr`:

```nora
// BEFORE:
pub fn (self: &PipelineLayout) drop() {
    sys.wgpuPipelineLayoutRelease(self.handle)
}

// AFTER:
pub fn (self: &PipelineLayout) drop() {
    if self.handle != none {
        sys.wgpuPipelineLayoutRelease(self.handle)
    }
}
```

Applied to: `PipelineLayout`, `BindGroupLayout`, `BindGroup`, `ComputePipeline`, `RenderPipeline`.

---

## Validation

After fix, `nora run --example cube` ran for thousands of frames continuously with no crash.

---

## Language Design Implications

This bug reveals a design convention gap: any Nora library wrapping opaque C/Rust handles
**must** implement null-safe drop functions, because the topology solver will always pre-drop
an initial default-constructed value (which has `handle: none`) before the first real assignment.

### Established Convention

> **Rule:** Any struct in `nora_wgpu` that holds a raw `ptr` handle to an external GPU
> resource MUST guard its `drop()` function with `if self.handle != none`.

### Compiler Enhancement Consideration

A future improvement: the topology solver could detect when the *initial value* of a variable
is a struct literal composed entirely of zero/none fields and elide the pre-drop for that
specific first assignment. This would be an optimization, not a correctness fix.

---

## Files Changed

| File | Change |
|------|--------|
| `nora_wgpu/src/core/pipeline.nr` | Added null-guards to 5 GPU resource drop functions |
