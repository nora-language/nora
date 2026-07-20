package main

pub fn main() {
    var device = alloc Device { data: 0.0 }
    
    // We should be able to instantiate a builder and use standalone 
    // method chains that discard their mutable lease returns without 
    // triggering a compiler error about dropping owned values!
    var pb = NewPipelineBuilder(#device)
    
    // Test the discard (previously caused "cannot discard owned value of type '&PipelineBuilder'")
    pb.WithShader().WithDepth()
}

pub type Device = struct {
    data: f64
}

pub type PipelineBuilder = struct {
    device: #Device,
    depth: i32
}

pub fn NewPipelineBuilder(dev: #Device) @PipelineBuilder {
    return alloc PipelineBuilder {
        device: dev,
        depth: 0
    }
}

pub fn (pb: &PipelineBuilder) WithShader() &PipelineBuilder {
    return pb
}

pub fn (pb: &PipelineBuilder) WithDepth() &PipelineBuilder {
    pb.depth = 1
    return pb
}
