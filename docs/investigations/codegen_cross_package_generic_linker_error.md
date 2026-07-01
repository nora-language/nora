# Investigation: Cross-Package Generic Monomorphization Linker Errors

## Status
Workaround Applied (Pending Compiler Fix)

**Date:** June 27, 2026
**Component:** `pkg/codegen` (Type-Erased Shared Monomorphization)

## Problem
When compiling a generic struct from the standard library (e.g., `collections.Vector[T]`) across package boundaries, the C compiler (`clang`) fails at the linker stage with `unresolved external symbol` errors if the type parameter `T` is a large struct instantiated by value (e.g., `wheel.WheelInfo[f64]`).

For example, using `collections.Vector[wheel.WheelInfo[f64]]` in the `raycast_vehicle` package results in the linker complaining about missing symbols like:
`collections_Vector_50cc9b4d_Push_173c28dc`
`collections_NewVector_a948419e`

## Reproduction
1. Create a package `pkg_a` defining a value struct: `pub type Data[T] = struct { val: T }`.
2. In `pkg_b`, instantiate a standard generic collection using that value struct: `var list = collections.NewVector[pkg_a.Data[f64]](4)`.
3. Call methods on the generic collection: `list.Push[pkg_a.Data[f64]](...)`.
4. Compile using `Nora run`.
5. The compiler successfully parses, type-checks, and transpiles to C, but fails at `clang` linkage with `fatal error LNK1120: unresolved externals`.

## Root Cause
The Nora compiler employs Type-Erased Shared Monomorphization in `pkg/codegen`. When a generic collection is instantiated with a **pointer-like argument** (`ptr`, `str`, `@T`), the compiler cleverly merges these into a single shared implementation (`_ptr` erased type) usually generated in `out_globals.c` or within the root `collections.o` cache.

However, when a generic collection is instantiated with a **value layout configuration** (a concrete struct), a specific concrete monomorphized variant must be generated. Due to an incremental compilation/caching bug in `pkg/codegen` across package boundaries, if the standard library `collections` package wasn't fully processed with this new concrete value type during its own compilation phase, the code generator fails to emit the concrete C functions (e.g., `collections_NewVector_a948419e`) into the calling package's output or `out_globals.c`. Consequently, the generated caller C code references a function that was never emitted.

## Fix (Workaround Applied)
To bypass this compiler linker bug in `nora_physics`:
1. Re-architected the affected generic collection to hold **pointers** instead of **values**. By changing the type from `Vector[wheel.WheelInfo[T]]` to an internally specialized array wrapper (`WheelVector[T]`) holding `@(@wheel.WheelInfo[T])[]`, we forced the code to bypass the cross-package generic `collections.Vector` boundary and avoid the missing linker symbols. 
2. Another workaround is clearing the runtime cache (`build\debug\runtime_cache`) to ensure the compiler doesn't blindly link a stale `collections.o` object file that lacks the necessary concrete methods.

## Validation
By implementing the localized array wrapper `WheelVector` inside the caller package and using explicit pointer references, the transpiler correctly resolved the symbols within the package scope. The C linker phase passed successfully without missing external dependencies, validating that the bug is localized to cross-package monomorphization emission for value structs.
