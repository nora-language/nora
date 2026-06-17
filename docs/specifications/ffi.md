# Foreign Function Interface (FFI)

## Overview

Nora compiles down to C11, making C interoperability completely transparent and native. The Foreign Function Interface (FFI) in Nora allows calling external C libraries directly, exporting Nora functions to C, and managing raw memory manually when crossing the boundary.

## Motivation

Systems programming languages must interoperate seamlessly with the existing ecosystem of C libraries (e.g., graphics, OS primitives, physics engines). Nora aims to make calling C functions as straightforward as calling native Nora functions, without requiring expensive runtime wrappers or complex bindings.

## Syntax & Core Concepts

### 1. The `extern` Keyword
To call a C function, declare it using the `extern` keyword. Nora will generate code expecting the linker to provide the implementation.

```nora
extern fn printf(format: ptr, ...) i32
extern fn nr_free_untracked(p: ptr)
```

### 2. The `export` Keyword
To allow a C program to call a Nora function, use `export`. This prevents Nora from name-mangling the function in the generated C code, preserving its exact name.

```nora
export fn MyNoraCallback(data: i32) {
    io.PrintLn("Called from C!")
}
```

### 3. The `ptr` Type
Nora uses the `ptr` keyword to represent untyped raw C pointers (`void*`). This type circumvents the Topological Lease Solver's tracking, allowing arbitrary memory manipulation but placing safety burdens squarely on the developer.

### 4. C Strings
C expects null-terminated strings (`char*`), whereas Nora strings contain length headers and data pointers. The `ffi` package provides conversion utilities:

```nora
import "ffi"

var my_str = "Hello C"
var c_str = ffi.CString(my_str)  // Generates a null-terminated ptr

// Reconstruct a Nora string from a C ptr
var nora_str = ffi.StringFromCString(c_str) 

// Note: You must free the CString manually to avoid leaks!
nr_free_untracked(c_str)
```

## Semantics & Pointer Ownership

Crossing the FFI boundary often requires transferring ownership from Nora's automatic RAII solver into manual C space, and vice-versa.

### `ffi.IntoRaw` & `ffi.FromRaw`

*   **`IntoRaw[T](@T)`**: Takes ownership of a Nora object, strips it of lease tracking, and returns a raw `ptr`. The solver will *not* insert a `drop()` for this object. C code must free it, or it must be returned to Nora.
*   **`FromRaw[T](ptr)`**: Reclaims ownership of a raw pointer back into Nora's lease solver. Nora will now automatically track and `drop()` the resource when the scope ends.

```nora
// 1. Allocate tracked resource
var res = alloc Resource { id: 42 }

// 2. Transfer ownership to C (Nora stops tracking it)
var raw_ptr = ffi.IntoRaw[Resource](@res)

// 3. Reclaim ownership (Nora starts tracking it again)
var reclaimed_res = ffi.FromRaw[Resource](raw_ptr)
// reclaimed_res will be safely dropped here at end of scope
```

## Lease Rules

1.  **Pinning:** When passing a Nora read-only borrow (`#`) to C, the solver must ensure the object stays alive. The `pin` keyword can be used to manually extend leases across complex FFI boundaries if necessary.
2.  **Untracked Memory:** Any memory returned by a standard C `malloc` (via `extern`) is completely untracked. You must `free` it manually, or wrap it in an `ffi.Owned[T]` smart pointer to opt it back into the lease solver.

## Full C-Closure Callback Interoperability

Nora provides predictable, 100% native interoperability with standard C callbacks by passing `export fn` identifiers directly. 

### Stateless Callbacks
For simple callbacks (e.g., C's `qsort`), you can pass an `export fn` as a standard C function pointer without any wrappers:

```nora
export fn my_cmp(a: ptr, b: ptr) i32 {
    return 0
}

fn main() {
    call_c_function(my_cmp)
}
```

### Stateful Callbacks (Trampoline Pattern)
When interfacing with stateful C APIs that accept a callback and a `void* user_data` (like `EnumWindows`), you can pass Nora closures (which are fat pointers containing captured environment state) by leveraging the Trampoline pattern and `ffi.FromRaw`/`ffi.IntoRaw`:

```nora
// 1. Define a struct to hold the closure fat-pointer
type EnumCallback = struct {
    cb: fn(ptr) i32
}

// 2. Define a stateless trampoline that reconstructs the closure from user_data
export fn my_trampoline(hwnd: ptr, lParam: ptr) i32 {
    // Reconstruct the closure environment and execute it
    var wrapper = ffi.FromRaw[EnumCallback](lParam)
    return wrapper.cb(hwnd)
}

fn main() {
    var count = 0
    var my_closure = fn (hwnd: ptr) i32 {
        count = count + 1
        return 1
    }
    
    // 3. Allocate the wrapper and transfer ownership to C
    var wrapper = alloc EnumCallback{ cb: my_closure }
    call_stateful_c(my_trampoline, ffi.IntoRaw[EnumCallback](wrapper))
}
```

> **Warning**: Ensure you manage the closure's lifecycle correctly. If the C function caches the callback asynchronously, the `wrapper` must stay alive in Nora space, or you must avoid dropping it when reconstructed in the trampoline.

## Examples

### FFI Call

```nora
import "ffi"

extern fn puts(str: ptr) i32

fn main() {
    var c_str = ffi.CString("Direct C Call")
    puts(c_str)
    
    // Remember to clean up!
}
```

## Errors & Diagnostics

*   **Leaking FFI memory:** `nr_mem_report()` will flag `ffi.CString` allocations that are never freed during `--debug-memory` execution.
*   **Linker Errors:** If you declare an `extern fn` but do not link the corresponding C library (via `--headers` or standard system libs), compilation will fail at the native CC step with an undefined symbol error.
