# Struct Array Field Layout Mismatch in FFI/GPU Contexts

## Status
**Resolved (Workaround Applied)** — Proper language fix pending.

---

## Problem

When a fixed-size array type (e.g., `f32[2]`, `f32[3]`, `f32[4]`) is used as a **field inside a struct**, the Nora compiler emits a **slice header** (a struct containing a pointer and a length, approximately 16 bytes) in the generated C code — not a packed C-style inline array (`float arr[N]`).

This is the correct behavior for standalone slice variables, but it is **incorrect and dangerous** when the struct is intended to be passed to C FFI, written to a GPU buffer, or used in any context requiring a specific binary memory layout.

### Root Cause

Nora's `T[N]` syntax is sugar for a heap-allocated slice. When used as a local variable, this works perfectly. However, when embedded as a struct field, the compiler stores a **fat pointer** (data pointer + length) — not inline data. The resulting struct layout differs from what C, WGPU, or any other native interface expects.

---

## Reproduction

### Failing Pattern

```nora
pub type Particle = struct {
    pos: f32[2],   // ❌ Emits a slice header (~16 bytes), NOT 8 bytes of packed f32
    vel: f32[2],   // ❌ Same issue
    color: f32[4]  // ❌ Same issue
}
```

**Expected C layout** (for WGSL `struct Particle { pos: vec2<f32>, vel: vec2<f32>, color: vec4<f32> }`):
```c
// Expected: 8 + 8 + 16 = 32 bytes
float px, py, vx, vy, cr, cg, cb, ca;
```

**Actual C layout** (what Nora emits):
```c
// Actual: slice_header(pos) + slice_header(vel) + slice_header(color)
// Each slice header is {void* data; int64_t len;} = 16 bytes
// Total: 48+ bytes of garbage from the GPU's perspective
```

### Observed Failure

The `nora_wgpu` instancing example (`examples/instancing/main.nr`) had completely black output because:
1. `Vertex { pos: f32[3], color: f32[3] }` emitted slice headers as fields.
2. When written to the WebGPU vertex buffer via `queue.WriteBuffer`, the buffer was filled with raw C pointers and lengths instead of actual float coordinates.
3. The GPU vertex shader received nonsensical positions, causing all triangles to be clipped or discarded.
4. There were **no compiler errors or runtime warnings** — the failure was completely silent.

---

## Fix (Workaround Applied)

Replace array fields in FFI/GPU structs with **flat scalar fields**:

```nora
// ✅ Correct — flat f32 fields emit packed C scalars
pub type Particle = struct {
    px: f32, py: f32,
    vx: f32, vy: f32,
    cr: f32, cg: f32, cb: f32, ca: f32
}

pub type Vertex = struct {
    x: f32, y: f32, z: f32,
    r: f32, g: f32, b: f32
}
```

---

## Recommended Solutions (Priority Order)

### 1. Compiler Diagnostic (High Priority — Quick Win)
Add a **semantic analysis warning** when `T[N]` is used as a struct field, since this almost always indicates the user expects a packed array but gets a slice header:

```
Warning: fixed-size array type `f32[2]` as struct field 'pos' emits a slice header
in the generated C layout and will NOT produce a packed inline array.
For FFI/GPU-compatible structs, use flat scalar fields or await `[repr(C)]` support.
  --> examples/instancing/main.nr:11:5
   11 |     pos: f32[2],
```

### 2. `std/math` Package — GPU-Compatible Vector Types (Medium-Term)
Add a `math` standard library with explicitly flat-field primitive types. This becomes the idiomatic Nora way to write GPU/FFI structs:

```nora
// std/math/vec.nr
pub type Vec2 = struct { x: f32, y: f32 }
pub type Vec3 = struct { x: f32, y: f32, z: f32 }
pub type Vec4 = struct { x: f32, y: f32, z: f32, w: f32 }
pub type Mat4 = struct {
    m00: f32, m01: f32, m02: f32, m03: f32,
    m10: f32, m11: f32, m12: f32, m13: f32,
    m20: f32, m21: f32, m22: f32, m23: f32,
    m30: f32, m31: f32, m32: f32, m33: f32
}
```

Usage:
```nora
import "math"

pub type Particle = struct {
    pos: math.Vec2,
    vel: math.Vec2,
    color: math.Vec4
}
```

### 3. `[repr(C)]` Struct Attribute (Long-Term)
Add a `[repr(C)]` attribute (mirroring Rust's `#[repr(C)]`) so that `T[N]` fields inside marked structs are emitted as inline C arrays:

```nora
[repr(C)]
pub type Vertex = struct {
    pos: f32[3],   // ✅ With repr(C): float pos[3];
    color: f32[3]  // ✅ With repr(C): float color[3];
}
```

---

## Affected Areas

- Any struct passed via `queue.WriteBuffer` to WebGPU (vertex/uniform/storage buffers)
- Any struct passed via raw FFI (`ffi.BorrowToRaw`, `ptr()` casts)
- Any struct serialized to binary (networking, file I/O, memory-mapped files)
- `nora_wgpu` vertex descriptors, uniform buffers, storage buffers

---

## References

- Fixed file: `nora_wgpu/examples/instancing/main.nr`
- Mesh buffer upload: `nora_wgpu/src/gfx/mesh.nr`
- Compiler codegen: `pkg/codegen/` (where `[repr(C)]` should be implemented)
