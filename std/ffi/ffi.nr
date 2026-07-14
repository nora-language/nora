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

extern fn nr_write_u8(p: ptr, val: u8)
extern fn nr_read_u8(p: ptr) u8
pub fn WriteU8(p: ptr, val: u8) { nr_write_u8(p, val) }
pub fn ReadU8(p: ptr) u8 { return nr_read_u8(p) }

extern fn nr_write_i8(p: ptr, val: i8)
extern fn nr_read_i8(p: ptr) i8
pub fn WriteI8(p: ptr, val: i8) { nr_write_i8(p, val) }
pub fn ReadI8(p: ptr) i8 { return nr_read_i8(p) }

extern fn nr_write_u16(p: ptr, val: u16)
extern fn nr_read_u16(p: ptr) u16
pub fn WriteU16(p: ptr, val: u16) { nr_write_u16(p, val) }
pub fn ReadU16(p: ptr) u16 { return nr_read_u16(p) }

extern fn nr_write_i16(p: ptr, val: i16)
extern fn nr_read_i16(p: ptr) i16
pub fn WriteI16(p: ptr, val: i16) { nr_write_i16(p, val) }
pub fn ReadI16(p: ptr) i16 { return nr_read_i16(p) }

extern fn nr_write_u32(p: ptr, val: u32)
extern fn nr_read_u32(p: ptr) u32
pub fn WriteU32(p: ptr, val: u32) { nr_write_u32(p, val) }
pub fn ReadU32(p: ptr) u32 { return nr_read_u32(p) }

extern fn nr_write_i32(p: ptr, val: i32)
extern fn nr_read_i32(p: ptr) i32
pub fn WriteI32(p: ptr, val: i32) { nr_write_i32(p, val) }
pub fn ReadI32(p: ptr) i32 { return nr_read_i32(p) }

extern fn nr_write_u64(p: ptr, val: u64)
extern fn nr_read_u64(p: ptr) u64
pub fn WriteU64(p: ptr, val: u64) { nr_write_u64(p, val) }
pub fn ReadU64(p: ptr) u64 { return nr_read_u64(p) }

extern fn nr_write_i64(p: ptr, val: i64)
extern fn nr_read_i64(p: ptr) i64
pub fn WriteI64(p: ptr, val: i64) { nr_write_i64(p, val) }
pub fn ReadI64(p: ptr) i64 { return nr_read_i64(p) }

extern fn nr_write_f32(p: ptr, val: f32)
extern fn nr_read_f32(p: ptr) f32
pub fn WriteF32(p: ptr, val: f32) { nr_write_f32(p, val) }
pub fn ReadF32(p: ptr) f32 { return nr_read_f32(p) }

extern fn nr_write_f64(p: ptr, val: f64)
extern fn nr_read_f64(p: ptr) f64
pub fn WriteF64(p: ptr, val: f64) { nr_write_f64(p, val) }
pub fn ReadF64(p: ptr) f64 { return nr_read_f64(p) }

extern fn nr_write_ptr(p: ptr, val: ptr)
extern fn nr_read_ptr(p: ptr) ptr
pub fn WritePtr(p: ptr, val: ptr) { nr_write_ptr(p, val) }
pub fn ReadPtr(p: ptr) ptr { return nr_read_ptr(p) }

// BorrowToRaw takes a read-only borrow and returns a raw ptr.
[intrinsic("borrow_to_raw")]
[NoEmit]
pub fn BorrowToRaw[T](val: #T) ptr {
    return ptr(val)
}

// MutBorrowToRaw takes a mutable borrow and returns a raw ptr.
[intrinsic("mut_borrow_to_raw")]
[NoEmit]
pub fn MutBorrowToRaw[T](val: &T) ptr {
    return ptr(val)
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
