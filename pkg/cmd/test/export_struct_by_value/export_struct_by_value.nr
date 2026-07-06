package main

import "io"

pub type StringView = struct {
    data: ptr
    length: i32
}

pub export fn wgpu_log_callback(level: i32, message: StringView, userdata: ptr) {
    if level == 1 && message.length == 0 && userdata == ptr(0) {
        io.PrintLn("log")
    }
}

fn main() {
    var sv = StringView{data: ptr(0), length: 0}
    wgpu_log_callback(1, sv, ptr(0))
}
