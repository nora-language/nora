# Large Array Stack Overflow in Nora `alloc T[N]` Codegen

## Status
**Completed** — The Nora compiler has been updated to unconditionally heap allocate `alloc T[N]` using `nr_malloc_debug` in the HIR codegen phase. The VLA stack overflow bug is resolved.

---

## Problem

When `alloc T[N]` is used in Nora source code and `N` is either a **runtime variable** or a **large compile-time constant**, the C11 codegen emits a C Variable Length Array (VLA) or a large fixed-size stack array respectively. For large element counts this causes an **immediate stack overflow** at program startup — before any user code runs.

### Observed Failure

```nora
// nora_wgpu/examples/instancing/main.nr
var num_particles = 100000
var p_data = alloc Particle[num_particles]  // Particle = 32 bytes
// → 100,000 × 32 = 3,200,000 bytes (3.2 MB) on the stack
```

**Result:** The binary compiles successfully via Clang but exits with code 1 immediately at runtime. No output is produced — not even the first `io.PrintLn` call — because the OS kills the process on OS thread stack creation before `main()` executes.

Default OS thread stack sizes:
- **Windows:** 1 MB (default), configurable up to ~8 MB via linker flags
- **Linux/macOS:** 8 MB (default)

3.2 MB easily overflows the Windows 1 MB default stack.

---

## Reproduction

```nora
// Any Nora program with a large local array:
var num_particles = 100000
var p_data = alloc Particle[num_particles]  // Particle is 32 bytes → 3.2MB on stack
```

Confirmed to crash with `num_particles = 100000`. Confirmed to work with `num_particles = 1000` (32KB — safe).

---

## Root Cause

### Codegen emits C VLAs / large stack arrays

In `pkg/hir/hir_codegen.go` (or equivalent), the `Alloca` HIR instruction for array types generates a C declaration of the form:

```c
// Generated C (example):
_Particle p_data[100000];  // VLA or large fixed array on the stack frame
```

For runtime-variable `N`, this is a C99 **Variable Length Array** which lives on the stack. For compile-time-constant large `N`, it is a plain fixed-size C array — also on the stack.

Neither is safe for large sizes. The issue is **not** that `alloc` should mean stack — the Nora language spec says `alloc` is the allocation keyword for both stack and heap. But the codegen currently does not distinguish small (stack-safe) from large (must-heap) allocations.

### Pointer arithmetic is missing

A secondary blocker discovered during investigation: Nora's `ptr` type does not support `+` arithmetic (`ptr + i32` is a type error). This means writing into a raw `GetMappedRange` pointer at byte offsets is not expressible in pure Nora without a native C helper. Combined with the VLA issue, there is no pure-Nora workaround for large GPU staging buffers today.

---

## Fix Options

### Option A — Compiler: Large-array threshold (recommended)

In `pkg/hir/hir_codegen.go`, detect when an `Alloca` for an array type would exceed a size threshold (e.g. **4096 bytes**) and emit a heap allocation instead:

```c
/* Small (N * sizeof(T) <= 4096 bytes) — stack VLA stays: */
_Particle p_data[3];

/* Large (N * sizeof(T) > 4096 bytes) — heap allocated: */
_Particle* p_data = (_Particle*)nr_malloc(sizeof(_Particle) * 100000);
/* Corresponding drop inserted by topology solver: nr_free(p_data) */
```

The Topological Lease Solver (`pkg/topology/solver.go`) must then insert the corresponding `PreDrop` / `Drop` for the heap-allocated array when it goes out of scope, exactly as it would for any other heap allocation.

**Pros:** Transparent to user code — no syntax change required.  
**Cons:** Requires careful integration with the lease solver to track the implicit heap allocation.

### Option B — `Box[T]` stdlib type (future language feature)

Add a `Box[T]` owned heap-allocated array type to `std/collections`:

```nora
// User code:
var p_data = Box.New[Particle](num_particles)
// p_data.Ptr() returns a raw ptr for FFI use
// Automatically freed when p_data goes out of scope
```

This is explicit and composable but requires a language/stdlib feature that does not exist yet.

### Option C — `ffi.PtrOffset` workaround

Expose a `ffi.PtrOffset(p: ptr, offset: i64) ptr` function that wraps `(char*)p + offset` in C:

```nora
// ffi/ffi.nr addition:
pub fn PtrOffset(p: ptr, offset: i64) ptr {
    return nr_ptr_offset(p, offset)
}
```

This enables writing to a `GetMappedRange` pointer directly without a CPU staging buffer. It unblocks the `mappedAtCreation` upload pattern for GPU buffers — but the VLA root cause in `alloc T[N]` remains and needs Option A or B to fully resolve.

---

## Validation Criteria

- [ ] Confirm the threshold-based codegen fix emits `malloc` for `alloc T[N]` when `N * sizeof(T) > 4096`
- [ ] Confirm the Lease Solver inserts `nr_free` drops for heap-allocated arrays
- [ ] Positive test: `var big = alloc u8[1000000]` should compile and not crash at startup
- [ ] Negative test: any double-free or use-after-free should be caught by `--debug-memory`
- [ ] Re-enable `num_particles = 100000` in `nora_wgpu/examples/instancing/main.nr` after fix

---

## Current Workaround

**Resolved.** The particle count in `nora_wgpu/examples/instancing/main.nr` has been successfully restored to `100000` as the underlying compiler issue was addressed.

---

## Affected Components

| Component | File | Notes |
|---|---|---|
| HIR Codegen | `pkg/hir/hir_codegen.go` | `Alloca` instruction must check size threshold |
| Topology Solver | `pkg/topology/solver.go` | Must track heap-alloca drops |
| FFI stdlib | `std/ffi/ffi.nr` | Add `PtrOffset` for raw pointer arithmetic |
| Collections stdlib | `std/collections/` | `Box[T]` heap-array type (Option B) |
| nora_wgpu | `nora_wgpu/examples/instancing/main.nr` | Re-enable 100k particles after fix |

---

## References

- Discovered during Phase 12 (GPU Instancing) development of `nora_wgpu`
- Date: 2026-07-10
- Related: WebGPU `mappedAtCreation` GPU buffer initialization pattern
