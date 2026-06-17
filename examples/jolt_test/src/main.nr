package main

import "io"

[include("src/wrapper.h")]
type Vec3 = struct {
    x: f32,
    y: f32,
    z: f32,
}

[include("src/wrapper.h")]
extern fn jolt_init() ptr

[include("src/wrapper.h")]
extern fn jolt_create_sphere(system: ptr, x: f32, y: f32, z: f32, radius: f32) ptr

[include("src/wrapper.h")]
extern fn jolt_create_floor(system: ptr) ptr

[include("src/wrapper.h")]
extern fn jolt_optimize_broadphase(system: ptr)

[include("src/wrapper.h")]
extern fn jolt_step(system: ptr, deltaTime: f32)

[include("src/wrapper.h")]
extern fn jolt_get_position(system: ptr, body: ptr) Vec3

[include("src/wrapper.h")]
extern fn jolt_destroy(system: ptr)

extern fn printf(format: str, ...) i32

fn main() {
    io.PrintLn("Initializing JoltPhysics via Custom C++ Wrapper...")
    
    var system = jolt_init()
    io.PrintLn("System initialized.")
    
    var _floor = jolt_create_floor(system)
    var sphere = jolt_create_sphere(system, f32(0.0), f32(10.0), f32(0.0), f32(1.0))
    jolt_optimize_broadphase(system)
    io.PrintLn("Created physics bodies and optimized broadphase.")
    
    var i = 0
    while i < 60 {
        jolt_step(system, f32(0.016))
        var pos = jolt_get_position(system, sphere)
        
        // Print every 10 steps to reduce terminal spam
        if i % 10 == 0 {
            printf("Step %d: Sphere Y Position = %f\n", i, f64(pos.y))
        }
        
        i = i + 1
    }
    
    jolt_destroy(system)
    io.PrintLn("Simulation complete!")
}
