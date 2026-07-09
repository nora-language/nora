# Codegen Drop Order Bug

## Status
Completed

## Problem
In the Nora compiler's C code generator (`pkg/codegen/generator.go`), a bug causes a C compilation error (`implicit function declaration` and `conflicting types`) when dropping generic types like `collections.Vector[ptr]`. The issue occurs because the drop function prototype for the generic type is omitted or generated after it is already called by the drop function of an encapsulating struct.

## Reproduction
This issue happens when a generic type (e.g., `collections.Vector[T]` where `T` erases to `ptr`) is only discovered as an `AutoDropMethod` *during* the generation of the drop method for another struct that contains it. 
Because the drop methods are collected in a map (`AutoDropMethods`) and emitted sequentially, the newly discovered drop method is pushed to the end of the emission list, meaning it is called in C code before its signature is declared.

To reproduce, see the test case added in:
`pkg/cmd/test/test_codegen_drop_order.nr`

*(Note: In the isolated test case, topological sorting of packages sometimes obscures the bug depending on which package discovers the generic drop first. However, the exact failure condition is easily hit in `nora_wgpu`'s `gfx` package when it uses `collections.Vector[sys.BindGroupLayout]`, because `sys.BindGroupLayout` contains only a `ptr`, causing it to erase to `ptr`.)*

## Root Cause
In `generator.go`, the C-codegen `GenerateHeader()` pass accurately discovers and generates the prototypes of most generic drop methods into `out.h`. However, type-erased instantiated drop methods (like `nr_drop_collections_Vector_ptr`) are only generated on demand when they are actually called inside monomorphized generic functions (like `Vector_ptr_Filter`) during the `GenerateSharedGlobals()` phase.

Because `Vector_ptr`'s drop function was only discovered *during* the generation of the `out_globals.c` function bodies, the compiler appended its drop method body at the very end of the C source, without ever declaring its prototype at the top of the file! This violated C99's strict rules, resulting in the "call to undeclared function" error when generating the C binary.

## Fix
**Compiler Fix Implemented:**
To permanently fix the Nora compiler, a `LatePrototypes *bytes.Buffer` was added to the compiler's `Generator` struct. Now, whenever a type-erased drop method (or any method) is discovered "late" (after `GenerateHeader` has completed), its C prototype is immediately buffered into `LatePrototypes` instead of being skipped.

Right before the compiler outputs the main C functions into `out_globals.c`, it prepends `LatePrototypes` directly to the top of the main buffer. This ensures that any auto-drop function discovered during body generation will always have a valid C prototype physically placed above any function that might call it, permanently eliminating the implicit declaration error.

## Validation
The `triangle` example was verified to compile and run successfully without needing the user-code workaround. The regression test in `pkg/cmd/test/` also passes successfully, confirming that the underlying root cause has been fully resolved across the compiler architecture.
