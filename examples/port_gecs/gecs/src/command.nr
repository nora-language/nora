package gecs

import "collections"

var DUMMY: i32 = 0

pub type ComponentBox[T] = struct {
    val: T
}

pub fn NewCommandBuffer() @CommandBuffer {
    return alloc CommandBuffer {
        commands: collections.NewVector[Command](64),
        boxes: collections.NewVector[any](16)
    }
}

pub fn (cb: &CommandBuffer) Spawn(w: &World) Entity {
    var e = w.reserve()
    var dummy_any: any = &DUMMY
    var cmd = Command {
        entity: Entity {
            id: e.id
        },
        cid: 0,
        comp_any: dummy_any,
        action: 4
    }
    cb.commands.Push[Command](@cmd)
    return e
}

pub fn (cb: &CommandBuffer) Despawn(e: #Entity) {
    var dummy_any: any = &DUMMY
    var cmd = Command {
        entity: Entity {
            id: e.id
        },
        cid: 0,
        comp_any: dummy_any,
        action: 3
    }
    cb.commands.Push[Command](@cmd)
}

pub fn (cb: &CommandBuffer) AddDeferred[T](e: #Entity, cid: i32, comp: T) {
    var val = comp
    var val_any: any = val
    var cmd = Command {
        entity: Entity {
            id: e.id
        },
        cid: cid,
        comp_any: val_any,
        action: 1
    }
    cb.commands.Push[Command](@cmd)
}

pub fn (cb: &CommandBuffer) Playback(w: &World) {
    io.Print("   [Playback] Starting playback of ")
    io.Print(io.FormatInt(cb.commands.size))
    io.PrintLn(" commands...")
    for i in 0..cb.commands.size {
        var cmd = &cb.commands.data[i]
        io.Print("   [Playback] Command ")
        io.Print(io.FormatInt(i))
        io.Print(": action = ")
        io.Print(io.FormatInt(cmd.action))
        io.Print(", entity ID = ")
        io.Print(io.FormatInt(cmd.entity.ID()))
        io.Print(", cid = ")
        io.Print(io.FormatInt(cmd.cid))
        io.PrintLn("")
        if (cmd.action == 4) {
            io.PrintLn("   [Playback] Spawning reserved...")
            w.spawnReserved(cmd.entity)
            io.PrintLn("   [Playback] Reserved spawned.")
        } else if (cmd.action == 3) {
            io.PrintLn("   [Playback] Despawning...")
            w.Despawn(cmd.entity)
            io.PrintLn("   [Playback] Despawned.")
        } else if (cmd.action == 1) {
            io.PrintLn("   [Playback] Adding component raw...")
            var comp = cmd.comp_any
            w.AddRaw(Entity {
                    id: cmd.entity.id
                }, cmd.cid, comp)
            var dummy_any: any = &DUMMY
            cmd.comp_any = dummy_any
            io.PrintLn("   [Playback] Component raw added.")
        }
    }
    cb.commands.size = 0
    cb.boxes.size = 0
}
