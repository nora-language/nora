---
title: FFI Borrow Dangling Pointer Bug
status: Resolved
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

# Fix
The fundamental problem was that `ffi.BorrowToRaw` and `ffi.MutBorrowToRaw` are generic functions, and NORA intrinsically optimizes read-only leases (`#`) of primitive types to pass-by-value in C to maximize performance. Passing a primitive by value strips it of its memory address, causing the generic function to return the value itself, not the address of the caller's stack variable. 

To resolve this cleanly without removing the primitive pass-by-value optimization, we extended NORA's compiler attribute system. 

1. **New Attributes:** We added `[intrinsic("borrow_to_raw")]` and `[intrinsic("mut_borrow_to_raw")]` to the `ffi.BorrowToRaw` and `ffi.MutBorrowToRaw` functions in `std/ffi/ffi.nr`. We also added `[NoEmit]` so the generic implementations are not emitted to C.
2. **Compiler Intercept:** During the HIR to C codegen phase (`pkg/codegen/hir_codegen.go`), the compiler checks for the `[intrinsic]` attribute. If it matches, the compiler bypasses the standard function call generation.
3. **Inline Address-Of:** The codegen inspects the argument. If the argument is passed by value in C (primitive), it takes the address of the operand inline and casts it to `void*`: `((void*)(&(argStr)))`. If it is already passed by pointer, it simply casts it: `((void*)(argStr))`.

This allows taking the raw pointer of `#cmd_handle` to correctly emit `&(cmd_handle)` in the generated C code, resolving the WebGPU panic while maintaining robust language metadata architecture.

# Validation
After applying this compiler fix, `wgpuQueueSubmit` receives a stable, valid pointer to the command buffer handle, completely eliminating the validation panic natively without the need for manual C wrappers.
