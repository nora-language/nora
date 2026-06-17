package ffi

// C library declarations preprocessed to Nora runtime counterparts
extern fn malloc(size: i32) ptr

extern fn free(p: ptr)

extern fn strlen(s: #str) i32

extern fn memcpy(dest: ptr, src: #str, n: i32) ptr

// Untracked C heap allocators for FFI transfers
extern fn nr_malloc_untracked(size: i32) ptr

extern fn nr_free_untracked(p: ptr)

// Custom runtime conversion function
extern fn nr_str_from_cstring(p: ptr) str

// IntoRaw takes ownership of a managed value and returns a raw, untracked ptr.
// Drop safety is statically enforced by the topology solver.
pub fn IntoRaw[T](val: @T) ptr {
    return val
}

// FromRaw takes a raw ptr and reclaims ownership, wrapping it in a managed @T lease.
pub fn FromRaw[T](p: ptr) @T {
    return p
}

// Malloc allocates dynamic memory tracked by Nora's debug leak detector.
pub fn Malloc(size: i32) ptr {
    return malloc(size)
}

// Free deallocates dynamic memory tracked by Nora's debug leak detector.
pub fn Free(p: ptr) {
    free(p)
}

// CString allocates a standard null-terminated C string on untracked heap.
// The caller is responsible for freeing this memory using standard C free() or nr_free_untracked.
pub fn CString(s: #str) ptr {
    var len = strlen(s)
    var p = nr_malloc_untracked(len + 1)
    if (p == none) {
        return none
    }
    memcpy(p, s, len + 1)
    return p
}

pub fn StrFromCString(p: ptr) str {
    return nr_str_from_cstring(p)
}

// StringFromCString constructs a safe, RAII-managed Nora string from a raw null-terminated C string pointer.
pub fn StringFromCString(p: ptr) str {
    return nr_str_from_cstring(p)
}

// Owned wraps an FFI-allocated raw pointer in a generic RAII smart pointer.
// When an Owned value goes out of scope, its custom drop() is executed to clean up the resource.
pub type Owned[T] = struct {
    raw: ptr
}

// Custom RAII destructor for Owned[T] that consumes raw and cleanly reclaims/drops the T resource.
pub fn (self: Owned[T]) drop() {
    if (self.raw == none) {
        return
    }
    // Reclaim ownership of T, transferring it to local variable `_val`
    var _val: @T = FromRaw[T](self.raw)
    // _val goes out of scope here and is automatically cleaned up by Nora's RAII drop system!
}
