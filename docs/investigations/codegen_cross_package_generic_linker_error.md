# Investigation: Cross-Package Generic Monomorphization Linker Errors

## Status
Completed

**Date:** June 27, 2026 (Fixed July 2026)
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

## Fix / Implementation
1. **Track Monomorphized Functions**: Added a `MonomorphizedFuncs` map to the `Generator` state in `generator.go`. During Step 2 (`CollectDefinitions`), any concrete generic functions created (like `collections_NewVector_a948419e`) are now registered in this map.
2. **Prevent Caching Bugs**: In `GeneratePackageCode`, added a check to skip generating these monomorphized functions into their generic template's source file (`out_pkg_collections.c`). This prevents the functions from being lost when the incremental compiler loads a previously cached `.o` object for library packages.
3. **Emit in Shared Globals**: Implemented `g.emitMonomorphizedFunctions()` and called it inside `GenerateSharedGlobals`. Because `out_globals.c` is dynamically constructed and hashed on every build, the C compiler will now correctly link every single monomorphized function without disrupting any package caches.

## Validation
- The internal `pkg/codegen` unit tests verify `MonomorphizedFuncs` properly buffer static arrays, spawn wrappers, string literals, and generic methods cleanly in `out_globals.c`.
- Full compiler integration tests successfully re-route generic linker dependencies through `out_globals.c` across packages without stale cache failures.
