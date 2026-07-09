# Nested Struct Auto-Drop Leak

## Status
Investigating

## Problem
When a struct contains owned fields (`@` pointers), the Nora compiler is expected to automatically generate an `AutoDropMethod` that recursively calls the `drop()` methods of its fields when the parent struct goes out of scope. However, during the implementation of `nora_wgpu`, we discovered a memory leak where `gfx.Texture`, `gfx.Mesh`, and `gfx.Material` structs were successfully deallocated, but their internal `@` fields (such as `@core.Texture`, `@core.TextureView`, `@core.Sampler`) were never dropped. 

Because the compiler failed to auto-drop the inner fields, the 8-byte pointer allocations for those inner objects leaked.

## Reproduction
To reproduce this issue, create a wrapper struct that takes ownership of another allocated object without defining a custom `drop()` method:

```nora
pub type Inner = struct {
    handle: ptr
}

pub fn (self: &Inner) drop() {
    // some cleanup
}

pub type Wrapper = struct {
    inner: @Inner
}

fn main() {
    var raw = alloc Inner { handle: none }
    var wrap = alloc Wrapper { inner: @raw }
    // Loop break or scope exit
    // Expected: `wrap` is dropped, which triggers `wrap.inner.drop()`, which cleans up `raw`.
    // Actual: `wrap` is dropped, but `raw` leaks (8 bytes).
}
```

## Root Cause
The root cause lies in how the Nora compiler's Topological Lease Solver and C-Codegen handle auto-generated drops. 

When a struct like `gfx.Texture` has no explicitly defined `drop()` method, the compiler's semantic phase correctly identifies it as needing an `AutoDropMethod` because it contains `@` fields. However, the generated C drop function `nr_drop_gfx_Texture` appears to be incomplete. It successfully frees the memory of the `gfx.Texture` allocation itself (the 24 bytes), but it fails to emit the recursive `drop()` calls for its `texture`, `view`, and `sampler` fields before the free occurs. 

This indicates that either the AST traversal in `requestAutoDrop` misses nested `@` fields, or the `emitAutoDropMethods` function in `generator.go` fails to iterate over the struct's layout and emit the necessary field cleanup instructions.

## Fix / Workaround
**Workaround Implemented:**
To bypass this compiler bug safely in the `nora_wgpu` project, explicit `drop()` methods were manually added to all wrapper structs (`gfx.Texture`, `gfx.Mesh`, and `gfx.Material`), explicitly un-referencing the `@` pointers. By manually defining the `drop()`, we override the faulty auto-generated drop logic.

```nora
pub fn (self: &Texture) drop() {
    var _tex = @self.texture
    var _view = @self.view
    var _samp = @self.sampler
}
```
*(Note: As discovered during this workaround, doing this requires a cache clear of the `build/debug/runtime_cache` to avoid linking errors, due to a separate caching bug where the compiler continues to look for the `nr_drop_...` symbol in dependent packages).*

**Required Compiler Fix:**
The `pkg/codegen/generator.go` (and potentially the semantic analyzer) must be updated. The logic that generates the body for `AutoDropMethods` must be amended to loop over all fields of the struct. For any field marked as an owned type (`@`), it must explicitly emit a call to that field type's drop function before freeing the struct's own memory.

## Validation
By implementing the manual `drop()` workaround, the active allocations dropped to `0` and the "Nora MEMORY LEAK REPORT" disappeared at runtime, proving that the leak was isolated purely to the auto-drop field traversal. A permanent regression test should be added to `pkg/cmd/test/` once the compiler fix is deployed.
