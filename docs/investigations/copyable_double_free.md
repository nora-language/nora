# Copyable Double-Free Bug

**Status:** Fixed

## Problem
The compiler allowed the `[copyable]` attribute to be attached to any struct, including structs that contained owned linear types (such as `@Vector[T]`, strings, or channels). Because `[copyable]` bypasses the Topological Lease Solver's move tracking and instead performs a bitwise C-level copy (`var a = b`), this allowed owned pointers to be duplicated. 

When both copies of the struct eventually went out of scope, the RAII engine emitted `Drop()` calls on both structs. This caused the underlying owned pointer to be freed twice, resulting in a fatal double-free memory corruption at runtime.

## Reproduction
```nora
pub type Vector[T] = struct {
    data: @T[],
    len: i32
}

// BUG: PointData contains an owned Vector but is marked copyable!
[copyable]
pub type PointData = struct {
    x: i32,
    owned_vec: @Vector[i32] 
}

fn trigger_double_free() {
    var a = alloc PointData{ ... }
    var b = a // Bitwise copy! Pointer duplicated.
    // End of scope: a is dropped (freed), b is dropped (freed again -> Crash!)
}
```

## Root Cause
The `[copyable]` attribute parser (in Pass 1 of `pkg/semantic/analyzer.go`) simply attached `IsCopyable = true` to the struct's definition without validating the semantic properties of the struct's fields. Because field types are not resolved until Pass 2, the validation was completely missing.

## Fix
In `pkg/semantic/analyzer.go`, during the second pass (Struct Field Resolution), we introduced a strict safety check. After all fields of a struct are resolved:
1. The compiler checks if the struct is marked `IsCopyable`.
2. It iterates over all fields and evaluates `types.IsOwnedType(field.Type)`.
3. If any field is an owned type, the compiler statically rejects the struct and aborts compilation.

## Validation
- A negative integration test was added at `pkg/cmd/test/attribute_copyable_invalid/fail_copyable_owned.nr`.
- The test runner confirmed that the semantic analyzer successfully intercepts the invalid attribute and issues a diagnostic error: `struct '%s' cannot be marked [copyable] because field '%s' is an owned type`.
- All other integration tests, including those utilizing `[copyable]` on primitive-only structs (like `Fixed64` in `nora_physics`), continue to compile correctly.
