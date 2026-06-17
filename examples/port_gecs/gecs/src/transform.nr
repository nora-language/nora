package gecs

import "collections"

pub fn (v: Vec3) Add(other: Vec3) Vec3 {
    return Vec3 {
        x: v.x + other.x,
        y: v.y + other.y,
        z: v.z + other.z
    }
}

pub type TransformSystem = struct {
    access: @Access
}

pub fn (s: #TransformSystem) GetAccess() #Access {
    return #s.access
}

pub fn NewTransformSystem() @TransformSystem {
    var reads = NewBitSet()
    reads.Set(5) // CID 5 for Transform
    reads.Set(3) // CID 3 for Parent

    var writes = NewBitSet()
    writes.Set(6) // CID 6 for GlobalTransform

    return alloc TransformSystem {
        access: alloc Access {
            reads: reads,
            writes: writes
        }
    }
}

fn GetParentGlobalPosition(world: #World, e: Entity) Vec3 {
    if (Has(world, Entity {
            id: e.id
        }, 3)) {
        var p_link = Get[Parent](world, Entity {
                id: e.id
            }, 3)
        var p_ent = Entity {
            id: p_link.id.id
        }
        var p_pos = Vec3 {
            x: 0.0,
            y: 0.0,
            z: 0.0
        }
        if (Has(world, Entity {
                id: p_ent.id
            }, 5)) {
            var p_trans = Get[Transform](world, Entity {
                    id: p_ent.id
                }, 5)
            p_pos = p_trans.position
        }
        var p_global = GetParentGlobalPosition(world, Entity {
                id: p_ent.id
            })
        return p_global.Add(p_pos)
    }
    return Vec3 {
        x: 0.0,
        y: 0.0,
        z: 0.0
    }
}

pub fn (_s: #TransformSystem) Update(world: #World, _cb: &CommandBuffer) {
    import "io"
    io.PrintLn("      [Update] starting...")
    var iter = world.archetypes.Iter[i64, @Archetype]()
    io.PrintLn("      [Update] iter created...")
    var entry_opt = iter.Next[i64, @Archetype]()
    io.PrintLn("      [Update] first entry fetched...")

    var loop = true
    while (loop) {
        match entry_opt {
            Some(entry) => {
                var arch: &Archetype = entry.value
                io.Print("      [Update] arch id: ")
                io.Print(io.FormatInt(i64(arch.id)))
                io.PrintLn("")

                var has_both = arch.mask.Has(5) && arch.mask.Has(6)
                if (has_both) {
                    var col_trans_opt = arch.columns.Get[i32, any](#5)
                    var col_global_opt = arch.columns.GetMut[i32, any](#6)

                    match col_trans_opt {
                        Some(col_trans) => {
                            match col_global_opt {
                                Some(col_global) => {
                                    var vec_trans: #collections.Vector[Transform] = collections.Vector[Transform](col_trans)
                                    var vec_global: &collections.Vector[GlobalTransform] = collections.Vector[GlobalTransform](col_global)
                                    var ents = #arch.entities

                                    for i in 0..ents.size {
                                        var e = ents.data[i]
                                        io.Print("         [Update] Entity ID: ")
                                        io.Print(io.FormatInt(i64(e.ID())))
                                        io.PrintLn("")

                                        var local = #vec_trans.data[i]
                                        var global = &vec_global.data[i]

                                        var has_parent = arch.mask.Has(3)
                                        io.Print("         [Update] has_parent: ")
                                        if (has_parent) {
                                            io.PrintLn("true")
                                            var p_link = Get[Parent](world, Entity {
                                                    id: e.id
                                                }, 3)
                                            io.Print("         [Update] parent ID: ")
                                            io.Print(io.FormatInt(i64(p_link.id.ID())))
                                            io.PrintLn("")
                                        } else {
                                            io.PrintLn("false")
                                        }

                                        var parent_pos = GetParentGlobalPosition(world, Entity {
                                                id: e.id
                                            })
                                        global.position = parent_pos.Add(local.position)
                                        io.PrintLn("         [Update] loop iteration done")
                                    }
                                }
                                None => {}
                            }
                        }
                        None => {}
                    }
                }

                entry_opt = iter.Next[i64, @Archetype]()
            }
            None => {
                loop = false
            }
        }
    }
}
