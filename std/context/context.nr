package context

import "sync"
import "string"

[shared]
pub type Context = struct {
    cancelled: @sync.AtomicBool,
    done_chan: chan[bool],
    err_msg: str
}

pub fn WithCancel() @Context {
    return alloc Context {
        cancelled: sync.NewAtomicBool(false),
        done_chan: make(chan[bool], 1),
        err_msg: ""
    }
}

pub fn (self: &Context) Cancel(err_msg: str) {
    if (self.cancelled.CAS(false, true)) {
        self.err_msg = err_msg
        self.done_chan <- true
    }
}

pub fn (self: #Context) IsCancelled() bool {
    return self.cancelled.Load()
}

pub fn (self: #Context) Err() str {
    if (self.IsCancelled()) {
        return string.Clone(#self.err_msg)
    }
    return ""
}

pub fn (self: #Context) Done() #chan[bool] {
    return #self.done_chan
}
