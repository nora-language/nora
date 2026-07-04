---
title: FFI Borrow Dangling Pointer Bug
status: Resolved (Workaround Applied)
date: 2026-07-04
tags: [ffi, codegen, webgpu]
---

# Problem
The WebGPU triangle example consistently crashed with `CommandBuffer[Id(1,0)] does not exist` inside `wgpuQueueSubmit` immediately after encoding the render pass. This was baffling because `CommandEncoder_Finish` was clearly succeeding and returning a valid ID, yet WebGPU was throwing a panic indicating the buffer ID couldn't be found.

# Root Cause
The root cause was traced to a code-generation bug in the Nora compiler concerning `ffi.BorrowToRaw[ptr]`.

In `core/command.nr`, `queue.Submit` was written as follows:
```nora
pub fn (self: &Queue) Submit(commands: &CommandBuffer) {
    var cmd_handle = commands.handle
    var ptr_to_handle = ffi.BorrowToRaw[ptr](#cmd_handle)
    sys.wgpuQueueSubmit(self.handle, 1, ptr_to_handle)
}
```

When generating C code for `ffi.BorrowToRaw[ptr](#cmd_handle)`, the compiler generated:
```c
void* ptr_to_handle = ffi_BorrowToRaw_ptr(_env_ptr, cmd_handle);
```

Because `ptr` is treated as a primitive type, the C code generator incorrectly passed `cmd_handle` **by value** instead of passing its memory address (`&cmd_handle`). Consequently, `ffi_BorrowToRaw_ptr` was returning the address of its *own local stack argument* rather than the caller's stack variable. When `ffi_BorrowToRaw_ptr` returned, its stack frame was popped, leaving a dangling pointer. `wgpuQueueSubmit` then read garbage memory from the stack, randomly interpreting overlapping data as `Id(1,0)`, and panicked when it couldn't find it in the WebGPU registry.

# Fix / Workaround
Until the Nora compiler's `BorrowToRaw` code-generator for primitive types is patched, we bypass this entirely for WebGPU by creating a thin C wrapper function:
```c
WGPU_EXPORT void nr_wgpuQueueSubmitSingle(WGPUQueue queue, WGPUCommandBuffer command) {
    WGPUCommandBuffer cmds[] = { command };
    wgpuQueueSubmit(queue, 1, cmds);
}
```
This safely bridges the FFI boundary by explicitly passing the single handle by value and creating the array natively inside C.

# Validation
After applying this C wrapper and modifying `Queue.Submit` to use it, `wgpuQueueSubmit` receives a stable, valid pointer to the command buffer handle, completely eliminating the validation panic.
