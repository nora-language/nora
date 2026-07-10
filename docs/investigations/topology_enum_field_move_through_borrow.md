# Investigation: Enum/Value Field Incorrectly Moved Through Mutable Borrow

## Status
**Status**: Completed — Compiler Fixed (July 2026)

## Problem

When a struct field of an enum (or other value type) is accessed through a **mutable borrow receiver** (`&T`) or **read-only borrow receiver** (`#T`) inside a struct literal, the topological lease solver incorrectly treats the access as an **ownership move**. This zeroes the field in the original struct at runtime without emitting any compile-time error or warning.

The consequence is silent data corruption: the field reads correctly on the first access but becomes `0` (the zero-value of the type) for all subsequent accesses within the same function call and in any function called after.

Additionally, primitive enums marked with `[repr("type")]` were incorrectly identified as linear owned types instead of Plain-Old-Data (POD), exacerbating the issue for enums specifically.

## Reproduction

The bug was first observed in `nora_wgpu/src/gfx/pass.nr` during Phase 10 (Automatic Depth Attachment). The `ForwardPass.Resize()` method receives `self` as `&ForwardPass` (mutable borrow) and uses `self.depth_format` (a `sys.WGPUTextureFormat` enum) in two struct literals.

```nora
pub type ForwardPass = struct {
    depth_format: sys.WGPUTextureFormat,
    // ...
}

// self is a MUTABLE BORROW — it does NOT own depth_format
pub fn (self: &ForwardPass) Resize(w: i32, h: i32) {
    var desc = alloc sys.TextureDescriptor {
        format: self.depth_format,  // ← solver treats this as a MOVE
        // ...
    }
    // After this line: self.depth_format == 0 (Undefined)!

    var view_desc = alloc sys.TextureViewDescriptor {
        format: self.depth_format,  // ← reads a zeroed field — wrong value!
        // ...
    }
}
```

**Generated C (incorrect before fix):**
```c
// First use — moves and zeroes the source field
.format = ({
    sys_WGPUTextureFormat _tmp = self->depth_format;
    memset(&self->depth_format, 0, sizeof(sys_WGPUTextureFormat)); // BUG
    _tmp;
});

// Second use — self->depth_format is now 0 (Undefined)
.format = ({
    sys_WGPUTextureFormat _tmp = self->depth_format; // reads 0!
    memset(&self->depth_format, 0, sizeof(sys_WGPUTextureFormat));
    _tmp;
});
```

**Expected C (correct post-fix):**
```c
.format = self->depth_format;  // simple copy — no destructive zero-wipe
```

## Root Cause

Two distinct root causes contributed to this issue:

1. **Primitive Enums Erroneously Treated as Linear Types**: In `pkg/types/types.go` (`IsOwnedType`), all Sum Types (enums) were categorized as owned, linear types that trigger automatic RAII drop semantics and move zero-wiping. Primitive enums (annotated with `[repr("type")]`) were not excluded from this check, causing the compiler to view copying an enum as a move.
2. **Missing Literal Inspections for Borrowed Moves**: In `pkg/topology/solver.go` (`isMoveOperationForSelector`), the static analysis failed to deeply inspect expressions on the right-hand side of struct, array, and map literals. Because it failed to inspect the literal fields, it missed cases where a field belonging to a borrowed context (`#T` or `&T`) was illegally moved into a newly allocated literal, thereby failing to emit the required compile-time error.

## Affected Compiler Components

- **`pkg/types/types.go`** — Determines whether a type follows linear move/drop semantics (`IsOwnedType`).
- **`pkg/topology/solver.go`** — Detects structural move operations (`isMoveOperationForSelector`) and enforces borrow checker rules.

## Fix

The compiler was patched to correctly enforce memory safety semantics for enums and borrowed contexts:

1. **Primitive Enum Bypass**: Modified `IsOwnedType` in `pkg/types/types.go` to explicitly unwrap `SumType` structures and return `false` if `st.IsPrimitiveEnum` is true. This ensures primitive enums are treated as trivial POD copies, bypassing the generation of `PreDrop`, `Drop`, and `memset(0)` instructions entirely.
2. **Literal Move Inspection**: Updated `isMoveOperationForSelector` in `pkg/topology/solver.go` to iterate through the fields/elements of `ast.StructLiteral`, `ast.ArrayLiteral`, and `ast.MapLiteral`. If the solver detects an implicit or explicit move of a field originating from a borrowed context, it now correctly flags it as a move operation.
3. **Semantic Errors**: With literal moves properly detected, the compiler correctly triggers the semantic error `"cannot move out of borrowed context"` if a developer attempts to move a linear type (like `str`) out of a borrow (`#T` or `&T`).

## Validation

- **Enum Move Verification**: Created `pkg/cmd/test/borrow_move_bug/borrow_move_bug.nr` to instantiate an `[repr("i32")]` enum, move it into a struct literal, and verify that the original value is not zero-wiped and continues executing safely without panic.
- **Borrowed Context Escape Verification**: Created `pkg/cmd/test/borrow_move_bug/borrow_move_bug2.nr` attempting to move an owned `str` out of a `#Pass` read-only struct receiver. Validated that the compilation gracefully fails with the expected error: `Error: cannot move out of borrowed context 'self'`.
- **System Integration**: The workaround snapshotting previously required in `nora_wgpu/src/gfx/pass.nr` is no longer necessary. The framework compiles successfully under the clean compiler.
