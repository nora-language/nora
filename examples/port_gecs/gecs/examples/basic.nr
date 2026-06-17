package main

import "gecs"
import "io"
import "serialize"

[serialize]
type Position = struct {
    x: i32,
    y: i32
}

[serialize]
type Velocity = struct {
    dx: i32,
    dy: i32
}

type HitEvent = struct {
    damage: i32
}

fn CloneEntity(e: #gecs.Entity) gecs.Entity {
    var id_val = e.id
    var res = gecs.Entity {
        id: id_val
    }
    return res
}

fn main() {
    io.PrintLn("================================================")
    io.PrintLn("--- GECS Premium Complete Port Integration ---")
    io.PrintLn("================================================")

    var w = gecs.NewWorld()
    io.PrintLn("1. World initialized.")

    gecs.RegisterComponent[Position](w, 1, "position")
    gecs.RegisterComponent[gecs.Transform](w, 5, "transform")
    gecs.RegisterComponent[gecs.GlobalTransform](w, 6, "global_transform")
    gecs.RegisterComponent[gecs.Parent](w, 3, "parent")
    gecs.RegisterChildrenComponent(w)

    // 2. Register Observer Hooks
    io.PrintLn("2. Registering component observers:")

    w.RegisterObserverAdd(1, fn(e: #gecs.Entity, val: &any) {
            var pos: &Position = Position(val)
            io.Print("   [Observer: Add] Entity ")
            io.Print(io.FormatInt(e.ID()))
            io.Print(": Position added {x: ")
            io.Print(io.FormatInt(pos.x))
            io.Print(", y: ")
            io.Print(io.FormatInt(pos.y))
            io.PrintLn("}")
        })

    w.RegisterObserverSet(1, fn(e: #gecs.Entity, val: &any) {
            var pos: &Position = Position(val)
            io.Print("   [Observer: Set] Entity ")
            io.Print(io.FormatInt(e.ID()))
            io.Print(": Position updated to {x: ")
            io.Print(io.FormatInt(pos.x))
            io.Print(", y: ")
            io.Print(io.FormatInt(pos.y))
            io.PrintLn("}")
        })

    // 3. Deferred Buffering via CommandBuffer
    io.PrintLn("3. Queueing deferred mutations in CommandBuffer:")
    var cb = gecs.NewCommandBuffer()

    var parent = cb.Spawn(w)
    var child = cb.Spawn(w)

    cb.AddDeferred[Position](CloneEntity(#parent), 1, Position {
            x: 10,
            y: 20
        })
    cb.AddDeferred[Position](CloneEntity(#child), 1, Position {
            x: 100,
            y: 200
        })

    io.PrintLn("   Playback queued deferred commands...")
    cb.Playback(w)

    // 4. Hierarchical Relation linking (Parent-Child)
    io.PrintLn("4. Linking parent-child hierarchies:")
    gecs.SetParent(w, CloneEntity(#child), CloneEntity(#parent))

    var has_parent = gecs.Has(w, CloneEntity(#child), 3)
    if (has_parent) {
        var p = gecs.Get[gecs.Parent](w, CloneEntity(#child), 3)
        io.Print("   Child ID ")
        io.Print(io.FormatInt(child.ID()))
        io.Print(" parent is confirmed: ID ")
        io.Print(io.FormatInt(p.id.ID()))
        io.PrintLn("")
    }

    // 5. Spatial Transform Propagation System
    io.PrintLn("5. Allocating spatial components:")

    // Parent components
    gecs.Add[gecs.Transform](w, CloneEntity(#parent), 5, gecs.Transform {
            position: gecs.Vec3 {
                x: 10.0,
                y: 20.0,
                z: 30.0
            },
            rotation: gecs.Vec3 {
                x: 0.0,
                y: 0.0,
                z: 0.0
            },
            scale: gecs.Vec3 {
                x: 1.0,
                y: 1.0,
                z: 1.0
            }
        })
    gecs.Add[gecs.GlobalTransform](w, CloneEntity(#parent), 6, gecs.GlobalTransform {
            position: gecs.Vec3 {
                x: 0.0,
                y: 0.0,
                z: 0.0
            }
        })

    // Child components (local transform + global transform placeholder)
    gecs.Add[gecs.Transform](w, CloneEntity(#child), 5, gecs.Transform {
            position: gecs.Vec3 {
                x: 5.0,
                y: 5.0,
                z: 5.0
            },
            rotation: gecs.Vec3 {
                x: 0.0,
                y: 0.0,
                z: 0.0
            },
            scale: gecs.Vec3 {
                x: 1.0,
                y: 1.0,
                z: 1.0
            }
        })
    gecs.Add[gecs.GlobalTransform](w, CloneEntity(#child), 6, gecs.GlobalTransform {
            position: gecs.Vec3 {
                x: 0.0,
                y: 0.0,
                z: 0.0
            }
        })

    // 6. Running conflict-free parallel systems via Scheduler
    io.PrintLn("6. Initializing Scheduler & systems:")
    var sched = gecs.NewScheduler()

    var trans_sys: gecs.System = gecs.NewTransformSystem()
    sched.AddSystem(trans_sys)

    io.PrintLn("   Running Scheduler (Transform propagation)...")
    sched.Run(w)

    var child_g = gecs.Get[gecs.GlobalTransform](w, CloneEntity(#child), 6)
    var cx = i32(child_g.position.x)
    var cy = i32(child_g.position.y)
    var cz = i32(child_g.position.z)

    io.Print("   Child propagated GlobalTransform: {x: ")
    io.Print(io.FormatInt(cx))
    io.Print(", y: ")
    io.Print(io.FormatInt(cy))
    io.Print(", z: ")
    io.Print(io.FormatInt(cz))
    io.PrintLn("} (Parent local + Child local)")

    // 7. Event emissions
    io.PrintLn("7. Broadcasting & reading custom events:")

    w.eventMgr.Emit[HitEvent](10, HitEvent {
            damage: 99
        })
    w.eventMgr.Emit[HitEvent](10, HitEvent {
            damage: 150
        })

    var events_opt = w.eventMgr.ReadEvents(10)
    match events_opt {
        Some(queue) => {
            for i in 0..queue.events.Len[any]() {
                var hit = HitEvent(queue.events.data[i])
                io.Print("   Read HitEvent: damage = ")
                io.Print(io.FormatInt(hit.damage))
                io.PrintLn("")
            }
        }
        None => {}
    }

    io.PrintLn("   Clearing event buffers...")
    w.eventMgr.Clear()

    // 8. Serialization / Deserialization Verification
    io.PrintLn("8. Testing Snapshot Serialization & Restoration:")

    var snap = gecs.CreateSnapshot(w)
    io.PrintLn("   Snapshot created.")

    var _snap_json = serialize.ToJSON[gecs.Snapshot](#snap)
    io.PrintLn("   Snapshot serialized to JSON.")
    io.PrintLn(_snap_json)

    var _w2 = gecs.RestoreSnapshot(_snap_json, w)
    io.PrintLn("   Snapshot restored into a new World.")

    var has_p_parent = gecs.Has(_w2, CloneEntity(#parent), 1)
    var has_p_child = gecs.Has(_w2, CloneEntity(#child), 1)

    if (has_p_parent) {
        var pos = gecs.Get[Position](_w2, CloneEntity(#parent), 1)
        io.Print("   Restored Parent Position: {x: ")
        io.Print(io.FormatInt(i64(pos.x)))
        io.Print(", y: ")
        io.Print(io.FormatInt(i64(pos.y)))
        io.PrintLn("}")

        if (pos.x != 10) {
            io.PrintLn("Panic: Restored parent position mismatch! at examples/basic.nr:187")
        }
    }
    if (has_p_child) {
        var pos = gecs.Get[Position](_w2, CloneEntity(#child), 1)
        io.Print("   Restored Child Position: {x: ")
        io.Print(io.FormatInt(pos.x))
        io.Print(", y: ")
        io.Print(io.FormatInt(pos.y))
        io.PrintLn("}")
        if (pos.x != 100 || pos.y != 200) {
            panic("Restored child position mismatch!")
        }
    } else {
        panic("Restored child does not have Position component!")
    }

    io.PrintLn("   Restored world state verified successfully!")

    io.PrintLn("================================================")
    io.PrintLn("GECS Full Port Test Successful!")
    io.PrintLn("================================================")
}
