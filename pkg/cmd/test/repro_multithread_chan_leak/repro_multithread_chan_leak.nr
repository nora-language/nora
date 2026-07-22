package main

import "io"
import "runtime"
import "time"

pub type CommandEncoderDescriptor = struct {
    label: ptr
}

pub type CommandBufferDescriptor = struct {
    label: ptr
}

pub type RenderPassColorAttachment = struct {
    view: ptr
}

pub type RenderPassDescriptor = struct {
    color_attach: ptr
}

pub type RenderPassEncoder = struct {
    handle: ptr
}

pub fn (self: &RenderPassEncoder) drop() {
    if self.handle != none {
        // mock drop
    }
}

pub type CommandBuffer = struct {
    handle: ptr
}

// Notice: In the un-fixed code, elements in array slices like `alloc CommandBuffer[4]`
// do NOT have their drop() called automatically when the slice goes out of scope!
pub fn (self: &CommandBuffer) drop() {
    if self.handle != none {
        // mock drop
    }
}

pub type CommandEncoder = struct {
    handle: ptr
}

pub fn (self: &CommandEncoder) BeginRenderPass(_descriptor: #RenderPassDescriptor) @RenderPassEncoder {
    return alloc RenderPassEncoder { handle: ptr(200) }
}

pub fn (self: &CommandEncoder) Finish(_descriptor: #CommandBufferDescriptor) @CommandBuffer {
    return alloc CommandBuffer { handle: ptr(300) }
}

pub fn (self: &CommandEncoder) drop() {
    if self.handle != none {
        // mock drop
    }
}

pub fn CreateCommandEncoder() @CommandEncoder {
    var _desc = alloc CommandEncoderDescriptor {
        label: none
    }
    return alloc CommandEncoder {
        handle: ptr(100)
    }
}

pub fn render_worker(id: i32, results: chan[@CommandBuffer]) {
    var encoder = CreateCommandEncoder()
    
    var _color_attach = alloc RenderPassColorAttachment[1]
    _color_attach[0] = RenderPassColorAttachment { view: ptr(1) }
    
    var _pass_desc = alloc RenderPassDescriptor { color_attach: ptr(_color_attach) }
    
    {
        var raw_pass = encoder.BeginRenderPass(#_pass_desc)
        pin raw_pass
    }
    
    var cb_desc = alloc CommandBufferDescriptor { label: none }
    var cmd_buf = encoder.Finish(#cb_desc)
    
    results <- @cmd_buf
}

pub fn main() {
    runtime.PinThread()
    io.PrintLn("=== Reproducing Multithreaded Memory Leak ===")
    io.PrintLn("Running continuous render loop (Press Ctrl+C to stop)...")

    var res0 = make(chan[@CommandBuffer], 1)
    var res1 = make(chan[@CommandBuffer], 1)
    var res2 = make(chan[@CommandBuffer], 1)
    var res3 = make(chan[@CommandBuffer], 1)

    var iter = 0
    while iter < 100 {
        // Per-frame heap allocation (reproducing main loop leak)
        var view_desc = alloc CommandBufferDescriptor { label: none }

        spawn render_worker(0, res0.clone())
        spawn render_worker(1, res1.clone())
        spawn render_worker(2, res2.clone())
        spawn render_worker(3, res3.clone())

        var h0 = <-res0
        var h1 = <-res1
        var h2 = <-res2
        var h3 = <-res3

        // Recreating `buffers = alloc CommandBuffer[4]` where array elements are un-dropped
        var buffers = alloc CommandBuffer[4]
        buffers[0] = @h0
        buffers[1] = @h1
        buffers[2] = @h2
        buffers[3] = @h3

        pin view_desc

        iter = iter + 1
        if iter % 1000 == 0 {
            io.PrintLn("Processed ${iter} frames...")
        }
        time.Sleep(#time.Millisecond())
    }
}
