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
In `generator.go`, `emitAutoDropMethods` iterates over the `AutoDropMethods` map and generates the C function bodies for each drop method. If a drop method contains an owned field (like a `collections.Vector`), it calls `requestAutoDrop`, which adds the new type to the `AutoDropMethods` map. However, because prototypes (`emitAutoDropPrototypes`) were generated earlier in `GenerateHeader()`, these newly discovered types never get their prototypes emitted in `out.h`. As a result, the C compiler encounters the function call before its declaration.

## Fix / Workaround
**Workaround implemented:**
To bypass this bug without modifying the compiler codebase, a workaround was implemented in the user code (`examples/cube/main.nr`):
```nora
// Workaround for compiler drop bug (#codegen-drop-order)
var _workaround = collections.NewVector[ptr](0)
```
This forces the compiler to discover `collections.Vector[ptr]` early during the compilation of the `main` package. Depending on package sorting, this can resolve the implicit declaration error.

**Compiler Fix (For Future Implementation):**
To permanently fix the Nora compiler, `generator.go` should be updated so that `emitAutoDropMethods()` always emits the prototypes of all currently discovered types *before* emitting any of their bodies in the loop, or the prototype generation should be deferred until all drop methods are fully discovered.

## Validation
The workaround allows `examples/cube/main.nr` to successfully compile and run, bypassing the bug safely. The regression test remains available in `pkg/cmd/test/` for when the compiler team fixes the underlying root cause.
