# Nested Struct Auto-Drop Leak

## Status
Resolved

## Problem
When a struct contains owned fields (`@` pointers), the Nora compiler is expected to automatically generate an `AutoDropMethod` that recursively calls the `drop()` methods of its fields when the parent struct goes out of scope. However, during the implementation of `nora_wgpu`, we discovered a memory leak where `gfx.Texture`, `gfx.Mesh`, and `gfx.Material` structs were successfully deallocated, but their internal `@` fields (such as `@core.Texture`, `@core.TextureView`, `@core.Sampler`) were never dropped. 

Because the compiler failed to auto-drop the inner fields, the 8-byte pointer allocations for those inner objects leaked.

## Reproduction
To reproduce this issue, create two structs across different packages that share the **exact same name**. One must have an explicit `drop` method, and the other must hold the first as an `@` pointer without its own explicit `drop` method.

**pkg_a/inner.nr**:
```nora
package pkg_a

pub type Inner = struct {
    handle: ptr
}

pub fn (self: &Inner) drop() {
    // Explicit drop method
}
```

**pkg_b/inner.nr**:
```nora
package pkg_b
import "pkg_a"

// Shares the same name ("Inner"), but has no manual drop method
pub type Inner = struct {
    value: @pkg_a.Inner
}
```

**main.nr**:
```nora
package main
import "pkg_a"
import "pkg_b"

fn main() {
    var raw = alloc pkg_a.Inner { handle: none }
    var wrap = alloc pkg_b.Inner { value: @raw }
    // Expected: `wrap` is auto-dropped, which recursively drops `wrap.value`.
    // Actual: Compiler binds `wrap` to `pkg_a.Inner`'s drop function due to the name collision! 
    // It runs `pkg_a.Inner`'s drop on `wrap` and permanently leaks the 8-byte `@pkg_a.Inner` field.
}
```

## Root Cause
The root cause was a **cross-package name collision** during the C-codegen phase.

When a struct like `gfx.Texture` has no explicitly defined `drop()` method, the compiler's semantic phase correctly identifies it as needing an `AutoDropMethod`. However, during method resolution, the compiler searches `g.SemanticInfo.MethodSymbols` for an existing explicit drop method via `getDropMethod()`.

In `pkg/codegen/generator.go`, `getDropMethod` attempts a direct lookup. When that fails (as it should for `gfx.Texture`), it falls back to a loop over all registered methods. Due to a flaw in pointer unwrapping, the loop used a dangerous string-matching check: `(k.Name() != "" && k.Name() == base.Name())`. 

Because both `core.Texture` (which *does* have a manual drop method) and `gfx.Texture` share the same struct name (`"Texture"`), the compiler incorrectly matched `gfx.Texture` to `core.Texture`'s explicit `drop()` method. This tricked the compiler into believing `gfx.Texture` already had a manual drop function, causing it to completely skip generating the `AutoDrop` recursive memory release instructions.

When a `gfx.Texture` went out of scope, the compiler blindly called `core_Texture_drop` on the `gfx.Texture` memory layout, which silently failed to release its nested `@` fields and permanently leaked them (8 bytes per pointer).

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

**Compiler Fix (Applied):**
The `pkg/codegen/generator.go` file was updated to eliminate the dangerous string-matching fallback in `getDropMethod` and `isDropMethodReceiverOwned`. The keys in `MethodSymbols` are pointer types (`&Texture`), so the fix properly unwraps the pointer first, and then relies entirely on the robust `types.Equals()` check (which enforces both struct structure and exact package boundaries):

```go
kBase := k
if pt, ok := k.(*types.PointerType); ok {
    kBase = pt.Base
}
if types.Equals(kBase, base) { ... }
```

## Validation
A standalone regression test was created in `pkg/cmd/test/repro_nested_struct_auto_drop_leak` using two `Inner` structs in separate packages to recreate the cross-package collision. With the fix applied, the compiler correctly avoids the name collision, generates the `AutoDrop` code, and Memory Leak tracking reports exactly 0 leaked bytes. The manual workarounds in `nora_wgpu` have since been removed.
