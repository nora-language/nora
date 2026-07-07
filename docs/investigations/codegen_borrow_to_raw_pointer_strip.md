# Compiler Investigation: `borrow_to_raw` strips Address-Of (`&`) on pointer types

**Status**: Root Cause Identified / Workaround Implemented

## Problem
When calling `wgpuQueueSubmit` via FFI in `nora_wgpu`, the program would panic inside the underlying Rust `wgpu-native` library during the very first iteration of the render loop:
```text
thread '<unnamed>' (9648) panicked at ...\src\storage.rs:135:46:
CommandBuffer[Id(1,0)] does not exist
```
The panic occurred because `wgpuQueueSubmit` requires a C array of `WGPUCommandBuffer` pointers, but was receiving an invalid memory address that happened to resolve to the ID of the command buffer rather than a pointer to an array containing the ID. 

## Reproduction
When attempting to pass the address of a pointer variable using `ffi.BorrowToRaw` or `ffi.MutBorrowToRaw`:
```nora
var cmd_handle: ptr = ...
sys.wgpuQueueSubmit(self.handle, 1, ffi.BorrowToRaw[ptr](&cmd_handle))
```
The expected C code generated should take the address of the pointer variable (i.e. returning a `void**` equivalent pointer):
```c
wgpuQueueSubmit(self->handle, 1, ((void*)(&cmd_handle)));
```
However, the Nora compiler was generating C code that completely stripped the `&` operator:
```c
wgpuQueueSubmit(self->handle, 1, ((void*)(cmd_handle)));
```

## Root Cause
The root cause lies in a faulty compiler optimization within the `hir_codegen.go` backend, specifically around the `borrow_to_raw` and `mut_borrow_to_raw` intrinsics. 

In `pkg/codegen/hir_codegen.go` (around line 700):
```go
if g.isPointerTypeInC(unwrapped) {
    if instOp, ok := arg.(*hir.InstOperand); ok {
        if addrOf, ok := instOp.Inst.(*hir.AddressOf); ok {
            argStr = g.hirOperandStr(addrOf.Val) // BUG: Strips the AddressOf node!
        }
    }
}
```
The codegen logic intentionally strips the `AddressOf` AST node if the underlying type is already considered a pointer in C (such as Nora's `ptr` type, which lowers to `void*`). Because of this, it erroneously passed the **value** of the pointer rather than the **address** of the pointer. 

When passed to C FFI functions that expect a pointer-to-a-pointer (like an array of handles), this causes the C library to dereference the raw handle ID directly as a memory address, leading to registry mismatches and crashes.

## Fix (Workaround)
To bypass the compiler bug in `nora_wgpu` without modifying the compiler itself, we wrapped the raw pointer in a struct. Because structs are passed by reference and are not strictly evaluated as "pointers in C" by this specific intrinsic block, the compiler correctly preserves the address.

```nora
pub fn (self: &Queue) Submit(commands: &CommandBuffer) {
    // `commands` is a pointer to a struct that contains exactly one `ptr` field. 
    // This perfectly mimics a C array of pointers (`void**`), bypassing the bug!
    sys.wgpuQueueSubmit(self.handle, 1, ffi.MutBorrowToRaw[CommandBuffer](commands))
    commands.handle = none
}
```

## Validation
By applying the workaround above, the `triangle` example immediately stopped panicking and correctly rendered the WebGPU context to the window on Windows x64. The C compiler generated the correct memory address reference for the `CommandBuffer` struct.
