package main

import "ffi"
import "io"

extern fn nr_free_untracked(p: ptr)

type Resource = struct {
    id: i32
}

pub fn (_self: Resource) drop() {
    io.PrintLn("RESOURCE DROPPED!")
}

fn test_into_from_raw() {
    io.PrintLn("--- Test IntoRaw & FromRaw ---")

    // 1. Allocate resource on heap
    var res = alloc Resource {
        id: 42
    }

    // 2. Transfer ownership to raw ptr
    io.PrintLn("Transferring ownership to raw pointer...")
    var raw = ffi.IntoRaw[Resource](@res)
    io.PrintLn("Resource transferred (should NOT drop yet)")

    // 3. Reclaim ownership
    io.PrintLn("Reclaiming ownership from raw pointer...")
    var _reclaimed = ffi.FromRaw[Resource](raw)
    io.PrintLn("Ownership reclaimed. Exiting scope (should drop now)...")
}

fn test_owned_smart_pointer() {
    io.PrintLn("--- Test Owned[T] Smart Pointer ---")

    var raw = ffi.IntoRaw[Resource](alloc Resource {
        id: 100
    })

    io.PrintLn("Wrapping raw pointer in Owned[Resource]...")
    {
        var _owned = ffi.Owned[Resource] {
            raw: raw
        }
        io.PrintLn("Owned wrapper created. Exiting block...")
    }
    io.PrintLn("Block exited (should have dropped now).")
}

fn test_cstring_roundtrip() {
    io.PrintLn("--- Test CString Roundtrip ---")

    var original = "Premium Nora FFI String!"
    io.PrintLn("Original string:")
    io.PrintLn(original)

    io.PrintLn("Creating CString...")
    var c_str = ffi.CString(original)

    io.PrintLn("Recreating Nora string from CString...")
    var nora_str = ffi.StringFromCString(c_str)
    io.PrintLn("Recreated string:")
    io.PrintLn(nora_str)

    // Clean up untracked CString
    nr_free_untracked(c_str)
}

fn main() {
    test_into_from_raw()
    test_owned_smart_pointer()
    test_cstring_roundtrip()
    io.PrintLn("All FFI ownership tests completed successfully!")
}
