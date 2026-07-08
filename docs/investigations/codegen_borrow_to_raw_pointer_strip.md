# Compiler Investigation: `borrow_to_raw` strips Address-Of (`&`) on pointer types

**Status**: Completed / Fixed

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

Previously, the codegen logic intentionally stripped the `AddressOf` AST node if the underlying type was considered a "pointer in C". This was originally added to ensure that slices and arrays (`i32[]`) passed to FFI were sent as native single pointers (`int*`) instead of double pointers (`int**`).

However, this broad condition (`g.isPointerTypeInC(unwrapped)`) also matched Nora's raw `ptr` type (which lowers to `void*`). Because of this, it erroneously passed the **value** of the pointer rather than the **address** of the pointer when `borrow_to_raw` was called on a raw pointer. 

When passed to C FFI functions that expect a pointer-to-a-pointer (like an array of handles), this caused the C library to dereference the raw handle ID directly as a memory address, leading to registry mismatches and crashes.

## Fix
The intrinsic handler in `pkg/codegen/hir_codegen.go` was updated to explicitly intercept the `AddressOf` instruction and intelligently decide whether to strip it.

The logic now correctly preserves the `AddressOf` operator for raw `ptr` types (thus guaranteeing `((void*)(&(%s)))`), while continuing to strip the `AddressOf` operator for slices, arrays, and strings (`i32[]`, `str`, etc.) so they pass smoothly as native single pointers.

```go
stripAddressOf := false
if g.isPointerTypeInC(unwrapped) && unwrapped.Name() != "ptr" {
    stripAddressOf = true
}

if stripAddressOf {
    return fmt.Sprintf("((void*)(%s))", valStr)
} else {
    return fmt.Sprintf("((void*)(&(%s)))", valStr)
}
```

## Validation
Two integration tests were added to the compiler test suite to lock in this behavior:
1. `codegen_borrow_to_raw_pointer_strip`: Asserts that `borrow_to_raw` on a `ptr` generates a double-pointer (`p_addr != p`).
2. `repro_ffi_borrow_double_pointer`: Asserts that `borrow_to_raw` on an array/slice (`i32[]`) correctly generates a single native pointer (`native_ptr == borrowed_ptr`).

Both tests successfully compile and pass, ensuring that FFI interactions with C libraries behave predictably for both arrays and arrays of pointers.
