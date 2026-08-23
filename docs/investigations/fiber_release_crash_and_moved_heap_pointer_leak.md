# Compiler Investigation: Fiber Release Crash & Moved Heap Pointer Cleanup

## Status
Resolved

## Problem
When compiling concurrent multi-fiber applications (such as the 3D Async Texture Streaming Showcase) in Release mode (`-O3`), the process crashed on Windows x64. Concurrently, memory leak integration tests (`repro_gecs_serialization_leak`, `repro_implicit_deref_leak`, `repro_multithread_chan_leak`, and `repro_vector_element_cleanup_pass`) required consistent lifecycle management of moved heap pointers.

## Reproduction
1. Compiling `bot_mecha_warrior_async_showcase` in Release mode (`..\nora.exe build -r --example bot_mecha_warrior_async_showcase`) and running on Windows x64 with 16 worker threads.
2. Executing `go test -v ./pkg/cmd/nora -run TestCompilerWithTestFolder/repro_gecs_serialization_leak`.

## Root Cause
1. **Windows Fiber Stack and Context Recycling**:
   - `fiber_wrapper` in `std/runtime/fiber.c` previously looped with `while (1)` when recycling fibers. Win32 fiber stack pointers (`RSP`) cannot be reset without creating a new fiber context, leading to stack depletion and instruction pointer jumping to `0x0`.
   - `info->is_main` in `scheduler_spawn` was set to `(old_count == 0)`, causing background fibers spawned when `g_active_fibers == 0` to be misclassified as the root main fiber, which prematurely terminated the worker scheduler upon fiber completion.
   - Non-volatile floating-point and vector registers (`XMM6`–`XMM15`) were not preserved across fiber context switches on Windows x64 without `FIBER_FLAG_FLOAT_SWITCH`.
2. **Moved Heap Pointer Disposal in Codegen**:
   - `cleanMovedHeapPointers` in `pkg/codegen/hir_codegen.go` failed to check if a function parameter in C was a pointer (`strings.HasSuffix(cType, "*")`), leading to premature `nr_free` on pointers moved into generic collections (such as `Vector.Push(@tex)`).
   - If the target in C is a struct passed by value (e.g. `comp: MyComponent`), the outer heap cell is dereferenced and must be freed, whereas if the target is a pointer in C, ownership is retained by the callee.

## Fix
1. **`std/runtime/fiber.c`**:
   - Updated `worker_loop` and `scheduler_run_loop` to delete terminated fiber handles via `DeleteFiber(info->handle)` upon termination.
   - Enforced `info->is_main = (name && strcmp(name, "main") == 0)`.
   - Used `CreateFiberEx` and `ConvertThreadToFiberEx` with `FIBER_FLAG_FLOAT_SWITCH`.
2. **`pkg/codegen/hir_codegen.go`**:
   - Refined `cleanMovedHeapPointers` to accurately check if the destination is a pointer in C (`!g.isPointerTypeInC(destType) && !strings.HasSuffix(cDest, "*") && !types.IsPointerLike(destType)`).
3. **`nora_engine/examples/bot_mecha_warrior_async_showcase/main.nr`**:
   - Guarded submesh drawing to only bind textures when all 4 PBR maps are staged (`sub_tex_mask[s_i] == 15 && sub_bg_handles[s_i] != none`).

## Validation
- All integration tests (`repro_gecs_serialization_leak`, `repro_implicit_deref_leak`, `repro_multithread_chan_leak`, and `repro_vector_element_cleanup_pass`) pass with 0 leaks.
- `bot_mecha_warrior_async_showcase` compiles and runs at 60 FPS in Release mode (`-O3`) on Windows x64.
