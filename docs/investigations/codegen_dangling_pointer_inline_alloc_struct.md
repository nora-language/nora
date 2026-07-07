# Investigation: Dangling Pointer on Inline Alloc in Struct Initialization

## Status
Completed

## Problem
In `nora_wgpu`, the triangle example was experiencing a periodic memory leak on the C backend (`wgpu-native`). Running Nora with `-debug-memory` reported 0 leaks on the Nora side, indicating that all Nora-managed objects were being properly freed by the Garbage Collector / Topological Lease Solver. However, memory consumption of the process continued to grow during the render loop.

## Reproduction
Inside the render loop in `examples/triangle/main.nr`, a `sys.RenderPassDescriptor` was initialized with an inline allocation for its `colorAttachments` field:

```nora
var pass_desc = alloc sys.RenderPassDescriptor {
    nextInChain: none,
    label: sys.StringView { data: none, length: 0 },
    colorAttachmentCount: 1,
    colorAttachments: alloc sys.RenderPassColorAttachment {
        view: view.handle,
        resolveTarget: none,
        loadOp: sys.WGPULoadOp.Clear,
        storeOp: sys.WGPUStoreOp.Store,
        clearValue: sys.Color { r: 0.1, g: 0.2, b: 0.3, a: 1.0 },
        depthSlice: -1
    },
    // ...
}
```

This caused the program's memory footprint to continuously increase.

## Root Cause
The root cause was a combination of Nora's scoping rules for temporaries and the expected pointer semantics of C-interop structs.

1. **Implicit Temporary Variable**: The `alloc sys.RenderPassColorAttachment` expression evaluates to a managed pointer (`@sys.RenderPassColorAttachment`). Since it is used directly as a value for the `colorAttachments` field, the Nora compiler implicitly stores this value in a hidden temporary variable.
2. **Ptr Coercion**: The `colorAttachments` field of `sys.RenderPassDescriptor` is typed as a raw C pointer (`ptr`). When the compiler assigns the `@` type to the `ptr` field, the raw memory address is written into the struct, but ownership is not transferred.
3. **Eager Drop Execution**: According to Nora's scoping rules, temporary variables created during statement execution have their lifetime bounded by the end of the statement. Therefore, at the closing brace `}` of the `pass_desc` struct initialization, the Topological Lease Solver immediately drops the temporary `@sys.RenderPassColorAttachment` and frees its memory.
4. **Dangling Pointer in FFI**: When `pass_desc` is later passed to `wgpuCommandEncoderBeginRenderPass`, its `colorAttachments` field is a dangling pointer pointing to reclaimed memory. Reading this garbage data caused `wgpu-native` to behave unpredictably, leading to memory leaks inside the C backend as it failed to properly process or clean up the corrupted descriptor.

## Fix
To prevent the topological solver from eagerly dropping the allocation, the inline allocation was extracted into a local variable bound to the scope of the render loop:

```nora
var color_attachment = alloc sys.RenderPassColorAttachment {
    nextInChain: none,
    view: view.handle,
    // ...
}

var pass_desc = alloc sys.RenderPassDescriptor {
    colorAttachments: ffi.BorrowToRaw[sys.RenderPassColorAttachment](#color_attachment),
    // ...
}
```

By assigning the allocation to `color_attachment`, the lifetime of the object is extended to the end of the block. The `BorrowToRaw` intrinsic correctly passes the memory address into the struct field, and the pointer remains valid when `pass_desc` is passed to the WebGPU C API.

*(Additionally, `device.Poll(1)` was added to the end of the render loop to guarantee that `wgpu-native` flushes its deferred destruction queues and reclaims C-side memory every frame.)*

## Validation
After applying the fix and running `..\nora.exe run --example triangle`, the periodic memory inflation stopped completely. The application remains perfectly stable with a flat memory footprint, validating that WebGPU is now receiving a valid descriptor and successfully reclaiming resources.
