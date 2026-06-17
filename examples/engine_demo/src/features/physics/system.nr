package physics

import "gecs"
import "io"
import "ecs"

pub type PhysicsSystem = struct {
    jolt_system: ptr,
    time: ptr, // pointer to core Time
}

pub fn NewPhysicsSystem(jolt_sys: ptr, t: ptr) PhysicsSystem {
    return PhysicsSystem {
        jolt_system: jolt_sys,
        time: t,
    }
}

pub fn (_self: &PhysicsSystem) Init(_w: &gecs.World) {
    io.PrintLn("Physics System Initialized")
}

pub fn (self: &PhysicsSystem) Update(w: &gecs.World) {
    var dt = 0.016 // fixed delta time for physics
    JoltStep(self.jolt_system, f32(dt))
    
    var query = gecs.NewQuery2[RigidBody, gecs.Transform](w, 2, 1)
    
    var update_fn = fn(entity: gecs.Entity, rb: &RigidBody, trans: &gecs.Transform) {
        var pos = JoltGetPosition(self.jolt_system, rb.jolt_body_ptr)
        trans.position.x = pos.x
        trans.position.y = pos.y
        trans.position.z = pos.z
    }
    
    query.Each[RigidBody, gecs.Transform](update_fn)
}
