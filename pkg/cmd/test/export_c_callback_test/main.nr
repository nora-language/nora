package main

import "ffi"
import "io"

extern fn call_stateless(cb: ptr)
extern fn call_stateful(cb: ptr, user_data: ptr)

var cmp_called_state = make(i32[], 1)
var global_state = make(i32[], 1)

// 1. Stateless callback test using qsort
export fn my_cmp(a: ptr, b: ptr) i32 {
    cmp_called_state[0] = cmp_called_state[0] + 1
    return 0
}


// 2. Stateful callback trampoline written entirely in Nora
type EnumCallback = struct {
    cb: fn(ptr) i32
}
// BOOL CALLBACK EnumWindowsProc(HWND hwnd, LPARAM lParam);
export fn my_enum_trampoline(hwnd: ptr, lParam: ptr) i32 {
    // Reconstruct the closure fat pointer
    var wrapper = ffi.FromRaw[EnumCallback](lParam)
    return wrapper.cb(hwnd)
}

fn main() {
    // --- Test 1: Stateless Callback ---
    // Pass export fn directly as a C function pointer
    call_stateless(my_cmp)
    
    if (cmp_called_state[0] == 0) {
        io.PrintLn("Failed stateless callback")
        return
    }

    // --- Test 2: Stateful Trampoline ---
    
    // A stateful Nora closure capturing state array
    var my_closure = fn (hwnd: ptr) i32 {
        global_state[0] = global_state[0] + 1
        // Return 0 to stop enumeration after 5 windows
        if (global_state[0] >= 5) {
            return 0
        }
        return 1
    }
    
    // Pass trampoline as callback, and closure fat pointer as userdata
    var wrapper = alloc EnumCallback{ cb: my_closure }
    call_stateful(my_enum_trampoline, ffi.IntoRaw[EnumCallback](wrapper))
    
    if (global_state[0] == 0) {
        io.PrintLn("Failed stateful callback")
        return
    }

    io.PrintLn("SUCCESS")
}
