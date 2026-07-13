package main
import "std/io"

pub type Particle = struct {
    pos: f32[3],
    vel: f32[3],
    color: f32[4]
}

pub fn main() {
    var p = alloc Particle {
        pos: [f32(0.0), f32(0.0), f32(0.0)],
        vel: [f32(0.0), f32(0.0), f32(0.0)],
        color: [f32(0.0), f32(0.0), f32(0.0), f32(0.0)]
    }
    io.PrintLn("Particle created")
}
