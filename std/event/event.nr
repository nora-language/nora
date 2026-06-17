package event

extern fn nr_fiber_current() ptr

extern fn nr_fiber_suspend()

extern fn nr_fiber_resume(f: ptr)

extern fn nr_store_ptr(dest: ptr, val: ptr)

extern fn nr_store_i32(dest: ptr, val: i32)

extern fn nr_load_i32(p: ptr) i32

extern fn nr_load_ptr(p: ptr) ptr

extern fn nr_cooperative_yield()

pub fn Current() ptr {
    return nr_fiber_current()
}

pub fn Suspend() {
    nr_fiber_suspend()
}

pub fn Resume(f: ptr) {
    nr_fiber_resume(f)
}

pub fn StorePtr(dest: ptr, val: ptr) {
    nr_store_ptr(dest, val)
}

pub fn StoreI32(dest: ptr, val: i32) {
    nr_store_i32(dest, val)
}

pub fn LoadI32(p: ptr) i32 {
    return nr_load_i32(p)
}

pub fn LoadPtr(p: ptr) ptr {
    return nr_load_ptr(p)
}

pub fn Yield() {
    nr_cooperative_yield()
}
