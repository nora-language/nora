# Investigation: GPU Buffer MapAsync Hang & FFI Memory Isolation

**Status**: Resolved
**Date**: July 14, 2026

## Problem
The `profiling` example application (and any WebGPU buffer mapping utilizing `MapSync`) would either:
1. Hang infinitely without throwing any errors ("Not Responding" window).
2. Crash intermittently during the next frame submission with a WebGPU Validation Error: `Buffer with '' label is still mapped`.

## Reproduction
1. Call `buffer.MapSync()` in `nora_wgpu` which allocates a sync flag and wraps `wgpuBufferMapAsync`.
2. Wait in a `while` loop for the callback to change the flag to `1`.
3. The flag is never updated, causing an infinite loop. 
4. If the loop is removed, the function returns early without unmapping the buffer, causing the next frame's `queue.Submit` to panic.

## Root Cause
The root cause was a combination of two distinct issues spanning across FFI boundaries and the native WebGPU specification:

### 1. FFI Callback Signature Misalignment
The `wgpu-native` C library recently updated the `wgpuBufferMapAsync` callback signature to include four arguments:
`void (*)(WGPUMapAsyncStatus status, WGPUStringView message, void *userdata1, void *userdata2)`

However, the Nora wrapper was still using a two-argument signature:
`fn _buffer_map_callback(status: i32, userdata: ptr)`

Because of the C-ABI (Application Binary Interface) calling convention, when WebGPU executed the callback, Nora interpreted the `message` string pointer (the 2nd argument) as the `userdata` pointer. Our callback then wrote the `status` code into the memory address of the string, corrupting the local stack but silently missing the actual `flag_ptr`. The `flag_ptr` remained `0`, causing the main fiber to spin infinitely.

### 2. Missing FFI Raw Memory Primitives
During early debugging, an attempt was made to use Nora's built-in `alloc` keyword to create the sync flag. However, passing Nora-managed memory across an asynchronous FFI boundary caused conflicts with Nora's static Topological Lease Solver, leading to premature drops and memory panics. 
To resolve this, we needed to use unmanaged C-heap memory (`ffi.Malloc`). However, Nora's `std/ffi` lacked the primitive capabilities to directly read or write typed values (like `i32`) to raw memory addresses, making it impossible to read the sync flag.

## Fix
1. **Callback Signature Update**: Corrected `_buffer_map_callback` in `nora_wgpu/src/core/buffer.nr` to accept all four arguments `(status: i32, message: sys.WGPUStringView, userdata1: ptr, userdata2: ptr)`, ensuring the ABI aligns perfectly.
2. **Synchronous Fallback**: Removed the manual spin loop and replaced it with a blocking `device.Poll(1)` combined with `sys.WGPUCallbackMode(0)`, which forces `wgpu-native` to resolve the async operation synchronously on the main thread.
3. **FFI Primitive Extensions**: Temporarily added `nr_write_i32` and `nr_read_i32` to the `std/runtime/nora_runtime.c` to allow raw memory manipulation. 

## Validation
- The `profiling` example now successfully resolves GPU timestamps without hanging.
- No `Buffer is still mapped` validation errors occur.
- **Follow-up Action**: The `std/ffi` package must be fully expanded to support reading and writing all primitive types (`u8`, `i8`, `u16`, `i16`, `u32`, `i32`, `u64`, `i64`, `f32`, `f64`, `ptr`) to complete the FFI feature set.
