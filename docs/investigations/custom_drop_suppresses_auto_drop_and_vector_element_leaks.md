# Custom Drop Method Suppressing Auto-Drop and Type-Erased Vector Element Field Leaks

## Status
Resolved / Workaround & Investigation Established

## Problem
During the development and testing of 3D rendering examples in `nora_wgpu` (`gltf_viewer`), we encountered two related memory leak behaviors stemming from architectural limitations in Nora's RAII drop insertion, collection handling, and shared generic monomorphization (`GEMINI.md` Section 2.C):

1. **Custom `drop()` on a Struct Suppresses Automatic Field Drop (`@T`)**:
   When a parent struct (`Model`) does not define a custom `drop()` method, the Topological Lease Solver automatically generates an `AutoDropMethod` for its owned fields (`@collections.Vector[Mesh]`, `@collections.Vector[PBRMaterial]`, etc.) when the struct goes out of scope. However, if an explicit user-defined `pub fn (self: &Model) drop()` method is declared, the compiler calls `Model.drop()` but completely stops emitting automatic recursive drop/free instructions for the struct's `@` fields (`model.meshes`, `model.materials`). Because generic collections like `Vector[T]` do not have their own explicit `drop()` method and rely on auto-dropping, defining `Model.drop()` causes the underlying `@Vector` allocations and their backing array buffers (`alloc T[c]`) to leak.

2. **Type-Erased Collection Slices (`@T[]` inside `Vector[T]`) Do Not Drop Elements or Their Inner `@` Fields When Nested Inside Structs**:
   As outlined in `GEMINI.md` Section 2.C (*Type-Erased Shared Monomorphization*), if a generic struct like `collections.Vector[T]` is instantiated with pointer-like or struct arguments (`Vector[Mesh]`), the compiler merges the implementations into a shared pointer-erased variant (`_ptr`). When `Vector[Mesh]` (`data: @Mesh[]`) goes out of scope inside a nested struct (`Model`), the shared auto-drop method for the type-erased vector only frees the backing memory of the `data` array (`@ptr[]`). Because the shared `_ptr` destructor does not retain concrete type information (`Mesh`) for the elements inside `data`, it does not iterate over the active elements (`0..size`) to invoke `Mesh.drop()` or recursively free owned `@` pointers (`vertex_buffer: @core.Buffer`) stored inside those array elements. Therefore, any struct stored inside a `Vector[T]` inside a parent struct that owns heap resources (`@core.Buffer`, `@core.BindGroup`) will permanently leak those inner allocations when the vector is dropped via `AutoDropMethod`.

## Reproduction
To reproduce this behavior, create a nested struct holding `@collections.Vector[T]` where `T` contains owned heap allocations (`@InnerBuffer`). When the parent struct is instantiated and drops out of scope, the outer vector and struct are freed, but the type-erased inner `@InnerBuffer` allocations inside the slice elements leak unless explicitly cleaned up before the vector drops.

**pkg/cmd/test/repro_vector_element_real_leak/main.nr**:
```nora
package main

import "collections"
import "io"

pub type InnerBuffer = struct {
    id: i32
}

pub fn (_self: &InnerBuffer) drop() {
    io.PrintLn("InnerBuffer dropped")
}

pub type MeshElement = struct {
    vb: @InnerBuffer
}

pub type ContainerModel = struct {
    meshes: @collections.Vector[MeshElement]
}

fn reproduce_vector_element_leak() {
    var vec = collections.NewVector[MeshElement](2)
    var buf = alloc InnerBuffer { id: 42 }
    var mesh = alloc MeshElement { vb: @buf }
    vec.Push(@mesh)
    
    var model = alloc ContainerModel { meshes: @vec }
    // When `model` goes out of scope, `model.meshes` (@collections.Vector[MeshElement]) is dropped by AutoDropMethod.
    // However, because Vector[T] uses shared monomorphization (`_ptr`), freeing vec.data (@MeshElement[]) only releases the array memory,
    // skipping element-level recursive RAII drop of `vb: @InnerBuffer`.
}

fn main() i32 {
    reproduce_vector_element_leak()
    return 0
}
```

## Root Cause
1. **Suppression of Field Auto-Drop**:
   In `pkg/topology/solver.go` and `pkg/codegen/generator.go`, when `getDropMethod()` locates an explicit user-defined `drop()` method on a struct, the compiler marks that struct as having manual drop handling and skips emitting `AutoDropMethod` recursively for its owned fields. The compiler assumes that the user's manual `drop()` implementation is solely responsible for all field teardown, but because owned fields (`@`) require explicit consumption moves (`@self.field`) to be freed on the stack, regular method calls inside `drop()` cannot easily trigger struct deallocation.

2. **Type-Erased Slice Deallocation in Shared Monomorphized Generics**:
   In `pkg/codegen/generator.go` (`eraseType`), generic structures and methods holding pointer or struct types are erased to `_ptr` implementations (`Vector_ptr`). When `AutoDropMethod` is generated for `Vector_ptr`, the drop routine emits `Free` on `self.data` (`@ptr[]`), but lacks the concrete monomorphized element type (`T`) or vtable required to generate a runtime loop (`for i = 0; i < size; i++`) to recursively invoke `AutoDropMethod` or explicit `drop()` methods on each `data[i]`.

## Fix & Workaround
Until the compiler frontend (`pkg/topology/solver.go` and `pkg/codegen`) is updated to:
1. Automatically emit field drop calls right after an explicit `drop()` method completes (or provide a `super.drop()` / `#auto_drop` directive), and
2. Ensure monomorphized or type-erased `Vector[T]` destructors receive element drop callbacks or iterate and drop elements when `T` contains `@` fields,

The established and robust pattern for complex resource hierarchies (`nora_wgpu`) is:
- **Avoid defining `drop()` on structs containing `Vector[T]` fields** (`Model`, `ModelNode`) so that the compiler's automatic RAII cleanup cleanly frees all `@collections.Vector` allocations without interference.
- **Implement explicit `CleanUp()` methods** (`Mesh.CleanUp()`, `Material.CleanUp()`, `Model.CleanUp()`) that iterate through collection elements and explicitly move inner `@` pointers (`var vb = @self.vertex_buffer`) onto the stack right before loop/application termination. When these local stack variables go out of scope at the end of the `CleanUp()` block, the topological lease solver automatically and cleanly frees the underlying WGPU and heap allocations (`src/core/device.nr:193`).

## Validation
Verified in `nora_wgpu` via `examples/gltf_viewer/main.nr`. By removing `Model.drop()` (preserving `Vector` auto-cleanup) and invoking `model.CleanUp()` right after the render loop breaks, both the collection slice arrays and the 18 bytes of inner `@core.Buffer` allocations (`src/core/device.nr:193`) are released with 0 leaked bytes reported by the runtime memory leak checker.
