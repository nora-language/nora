package nursery

import "string"

[shared]
pub type Context = struct {
    task_count: i32,
    completion_chan: chan[str],
    first_err: str
}

pub fn New() Context {
    return Context {
        task_count: 0,
        completion_chan: make(chan[str], 1000),
        first_err: ""
    }
}

pub fn (self: &Context) start_task() {
    self.task_count = self.task_count + 1
}

pub fn (self: #Context) done(err_msg: str) {
    self.completion_chan <- err_msg
}

pub fn (self: &Context) wait() Result[bool, str] {
    var has_error = false

    var i = 0
    while i < self.task_count {
        var err_msg = <-self.completion_chan
        if (err_msg != "") {
            if (!has_error) {
                self.first_err = err_msg
                has_error = true
            }
        }
        i = i + 1
    }

    if (has_error) {
        return Err[bool, str](string.Clone(#self.first_err))
    }
    return Ok[bool, str](true)
}
