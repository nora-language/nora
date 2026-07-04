package runtime

extern fn nr_fiber_pin_thread()
extern fn nr_fiber_unpin_thread()

pub fn PinThread() {
    nr_fiber_pin_thread()
}

pub fn UnpinThread() {
    nr_fiber_unpin_thread()
}

extern fn nr_cooperative_yield()

pub fn Yield() {
    nr_cooperative_yield()
}
