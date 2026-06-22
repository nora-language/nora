# Investigation: Generic Interface Codegen & Parser Limitations

**Status:** Completed
**Date:** 2026-06-22

## Problem Statement
While implementing the `nora_physics` 3D physics library using strict generic typing for `f32`/`f64` support, we encountered a series of hard limitations in the current Nora compiler's Parser and C11 Codegen backend. 

These limitations specifically surround how the compiler handles polymorphism, interfaces, and function pointers.

## Discovered Limitations

### 1. Generic Interface C-Codegen Omission
**Issue:** The semantic analyzer correctly processes generic interfaces (e.g., `pub type Shape[T] = interface`). However, when casting a generic struct (like `Sphere[f64]`) to this interface, the C-codegen backend fails to emit the C-struct definition for the interface in `out.h`.
**Error:** `error: unknown type name 'shape_Shape_536804a2'`
**Details:** The codegen successfully generates the vtables and function signatures using the type name, but the underlying `typedef struct { void* data; void* vtable; }` is skipped entirely.

### 2. Strict Semantic Method Resolution
**Issue:** Attempting to bypass interfaces using purely structural templates (e.g., `pub fn Intersect[T, ShapeA](shapeA: #ShapeA)`) fails.
**Error:** `type parameter 'ShapeA' has no field or method 'Support'`
**Details:** Unlike C++ templates which lazily resolve methods during monomorphization, Nora's semantic frontend strictly requires type bounds (which currently don't exist structurally) and rejects method calls on unconstrained generic parameters.

### 3. Parser Corruption on Inline Function Pointers
**Issue:** Attempting to use inline generic function pointers as an alternative to interfaces corrupts the parser.
**Code:** `supportA: fn(c: #Ctx, d: #vector.Vector3[T]) @vector.Vector3[T]`
**Error:** `undefined identifier` and `wrong number of arguments` in subsequent scopes.
**Details:** The parser fails to properly tokenize the `@` pointer return type when nested inside a function argument list, leading to downstream AST corruption.

### 4. Unsupported Generic Type Aliases
**Issue:** Attempting to alias the function pointer to bypass the inline parser bug also fails.
**Code:** `pub type SupportFunc[T, Ctx] = fn(...)`
**Error:** `undefined type: 'SupportFunc'`
**Details:** Generic type aliases are either entirely unsupported or bugged in the current semantic pass.

## Resolution & Workaround
To preserve the `f32/f64` generic architecture requested by the user, we will utilize the `native` C-interop block in `nora.yaml`. We can manually define the missing generic interface structs (e.g., `shape_Shape_536804a2`) inside a custom C-header (`hack.h`) and link it during compilation. This cleanly patches the codegen omission without polluting the pure Nora `.nr` source code!
