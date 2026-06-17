# Full C-Closure Callback Interoperability

## Overview
Nora provides a predictable, 100% native way to interface with standard C callbacks. The Nora compiler correctly omits the `_env_ptr` injection for `export fn` signatures, allowing Nora functions to act as raw, standard C function pointers.

When interacting with C APIs, passing stateful closures directly via FFI can lead to register corruption (because Nora closures are fat pointers `nr_closure_t`, while C expects an 8-byte function pointer). To solve this, we use the **Trampoline Pattern** combined with `#define` macros or `extern fn` casts.

## 1. Stateless Callbacks

For simple callbacks like C's `qsort` or WebGPU uncaptured error callbacks that do not require captured state, you can directly pass an `export fn` to an `extern fn` argument of type `ptr`.

```nora
// In Nora
export fn error_stateless_cb(device: ptr, err_type: i32, message: WGPUStringView, ud1: ptr, ud2: ptr) {
    io.PrintLn("Stateless Error Callback Triggered!")
}

fn main() {
    // To pass the export fn to C without fat-pointer wrapping, 
    // we simply cast it to a raw pointer.
    var err_cb = ptr(error_stateless_cb)
}
```

## 2. Stateful Callbacks (The Trampoline Pattern)

When you need to pass a stateful closure (which captures state) to a C API, you must use the Trampoline pattern with explicit `ffi` lifecycle control:

1. **Define the Trampoline**: An `export fn` that receives the raw parameters from C.
2. **Box the Closure**: Store your closure inside a struct allocated on the heap.
3. **Pass via IntoRaw**: Convert the struct to a raw pointer using `ffi.IntoRaw` and pass it as the `userdata` argument to C.
4. **Reconstruct via FromRaw**: Inside the trampoline, reconstruct the closure fat pointer with `ffi.FromRaw` using the `userdata` parameter and execute it!

### Example

```nora
// 1. Define the callback signature struct
type AdapterCallback = struct { cb: fn(i32, ptr, WGPUStringView) }

// 2. Define the Trampoline (matches C signature)
export fn adapter_trampoline(status: i32, adapter: ptr, message: WGPUStringView, ud1: ptr, ud2: ptr) {
    if (status != 1) { // 1 = WGPURequestAdapterStatus_Success
        io.PrintLn("Adapter Request Failed!")
    }
    
    // 3. Reconstruct the closure from userdata
    var wrapper = ffi.FromRaw[AdapterCallback](ud1)
    wrapper.cb(status, adapter, message)
}

fn main() {
    // Our stateful closure
    var on_adapter = fn (status: i32, adapter: ptr, msg: WGPUStringView) {
        io.PrintLn("Adapter received!")
        // Can access local variables here!
    }
    
    // Allocate the wrapper and pass it via userdata
    var adapter_wrapper = alloc AdapterCallback { cb: on_adapter }
    
    // Pass to C
    var adapter_info = WGPURequestAdapterCallbackInfo {
        callback: ptr(adapter_trampoline),
        userdata1: ffi.IntoRaw[AdapterCallback](adapter_wrapper), // Pass state as userdata
        // ...
    }
}
```

## Note on C Source Linking (`nora.yaml` vs `[c_source]`)

Currently, Nora **does not** utilize a `[c_source("...")]` inline attribute for linking external C files.

To cleanly link C implementations (like `native.c` containing macros or helper functions), you define them in the project's **`nora.yaml`** manifest file under `native.source_files`:

```yaml
name: my_project
# ...
native:
  compiler: clang
  source_files: ["src/native.c"]
```

This ensures the C source is compiled as a separate translation unit and correctly linked against the generated Nora outputs without confusing inline attributes.
