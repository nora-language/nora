package main

pub type Particle = struct {
    pos: f32[3],
    vel: f32[3],
    color: f32[4]
}

pub fn main() {
    var p = alloc Particle {
        pos: alloc f32[3],
        vel: alloc f32[3],
        color: alloc f32[4]
    }
}
