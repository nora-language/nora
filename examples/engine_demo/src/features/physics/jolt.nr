package physics

pub type Vec3 = struct {
    x: f32,
    y: f32,
    z: f32,
}

[include("../jolt_test/src/wrapper.h")]
extern fn jolt_init() ptr

[include("../jolt_test/src/wrapper.h")]
extern fn jolt_create_sphere(system: ptr, x: f32, y: f32, z: f32, radius: f32) ptr

[include("../jolt_test/src/wrapper.h")]
extern fn jolt_create_floor(system: ptr) ptr

[include("../jolt_test/src/wrapper.h")]
extern fn jolt_optimize_broadphase(system: ptr)

[include("../jolt_test/src/wrapper.h")]
extern fn jolt_step(system: ptr, deltaTime: f32)

[include("../jolt_test/src/wrapper.h")]
extern fn jolt_get_position(system: ptr, body: ptr) Vec3

[include("../jolt_test/src/wrapper.h")]
extern fn jolt_destroy(system: ptr)

// Public Wrappers

pub fn JoltInit() ptr {
    return jolt_init()
}

pub fn JoltCreateSphere(system: ptr, x: f32, y: f32, z: f32, radius: f32) ptr {
    return jolt_create_sphere(system, x, y, z, radius)
}

pub fn JoltCreateFloor(system: ptr) ptr {
    return jolt_create_floor(system)
}

pub fn JoltOptimizeBroadphase(system: ptr) {
    jolt_optimize_broadphase(system)
}

pub fn JoltStep(system: ptr, deltaTime: f32) {
    jolt_step(system, deltaTime)
}

pub fn JoltGetPosition(system: ptr, body: ptr) Vec3 {
    return jolt_get_position(system, body)
}

pub fn JoltDestroy(system: ptr) {
    jolt_destroy(system)
}
