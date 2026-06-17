package main

import "ffi"
import "io"
import "gecs"
import "runtime"
import "core"
import "ecs"
import "physics"
import "rendering"

pub fn main() {
    runtime.PinThread()
    io.PrintLn("Starting Nora Engine Demo...")

    // 1. Initialize Engine & Time Core
    var engine = core.NewEngine()
    var time = core.NewTime()

    // 2. Initialize Window
    var _title = ffi.CString("Nora Feature Engine Demo")
    var win = rendering.make_window(800, 600, _title)

    // 3. Initialize GECS
    var ecs_wrap = ecs.NewEcsWrapper()
    gecs.RegisterComponent[gecs.Transform](#ecs_wrap.world, 1, "transform")
    gecs.RegisterUnserializableComponent[physics.RigidBody](#ecs_wrap.world, 2, "rigidbody")
    gecs.RegisterUnserializableComponent[rendering.MeshRenderer](#ecs_wrap.world, 3, "meshrenderer")

    // 4. Initialize Physics (Jolt)
    var _jolt_sys = physics.JoltInit()
    var phys_sys = physics.NewPhysicsSystem(_jolt_sys, ptr(#time))

    // 5. Initialize Rendering (WebGPU)
    var render_sys = rendering.NewRenderSystem(win.handle)
    render_sys.Init(#ecs_wrap.world)

    // 6. Setup Scene
    // Floor
    var _floor_jolt = physics.JoltCreateFloor(_jolt_sys)
    
    // Falling Sphere
    var _sphere_jolt = physics.JoltCreateSphere(_jolt_sys, f32(0.0), f32(10.0), f32(0.0), f32(1.0))
    physics.JoltOptimizeBroadphase(_jolt_sys)

    var _sphere_ent = ecs_wrap.world.Spawn()
    gecs.Add[gecs.Transform](#ecs_wrap.world, ecs.CloneEntity(#_sphere_ent), 1, gecs.Transform {
        position: gecs.Vec3 { x: 0.0, y: 10.0, z: 0.0 },
        rotation: gecs.Vec3 { x: 0.0, y: 0.0, z: 0.0 },
        scale: gecs.Vec3 { x: 3.0, y: 3.0, z: 3.0 }
    })
    gecs.Add[physics.RigidBody](#ecs_wrap.world, ecs.CloneEntity(#_sphere_ent), 2, physics.RigidBody {
        jolt_body_ptr: _sphere_jolt
    })

    var _mesh = render_sys.CreateMeshRenderer()
    gecs.Add[rendering.MeshRenderer](#ecs_wrap.world, ecs.CloneEntity(#_sphere_ent), 3, @_mesh)

    io.PrintLn("Scene Setup Complete. Entering Main Loop...")

    // 7. Engine Loop
    while engine.IsRunning() != 0 {
        if (win.should_close()) {
            break
        }

        time.Tick(f32(0.016))

        // Update Physics & sync Transform
        phys_sys.Update(#ecs_wrap.world)
        ecs_wrap.Step()
        
        render_sys.Update(#ecs_wrap.world)
        
        win.poll_events()
    }

    // 8. Cleanup
    physics.JoltDestroy(_jolt_sys)
    win.drop()
    io.PrintLn("Engine Shutdown.")
}
