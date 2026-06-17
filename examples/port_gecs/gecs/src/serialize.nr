package gecs

import "collections"
import "io"
import "json"
import "serialize"
import "string"

pub fn CreateSnapshot(w: #World) @Snapshot {
    var snap_entities = collections.NewVector[EntityRecordRef](w.entities.size)
    for i in 1..w.entities.size {
        var record = #w.entities.data[i]
        if (record.archetype != 0) {
            snap_entities.Push[EntityRecordRef](EntityRecordRef {
                    id: i,
                    gen: record.gen
                })
        }
    }

    var snap_archetypes = collections.NewVector[ArchSnapshot](w.archetypes.Len[i64, @Archetype]())

    var iter = w.archetypes.Iter[i64, @Archetype]()
    var entry_opt = iter.Next[i64, @Archetype]()
    var loop = true
    while (loop) {
        match entry_opt {
            Some(entry) => {
                var arch: #Archetype = entry.value
                if (arch.entities.size > 0) {
                    var a_mask = collections.NewVector[str](arch.columnMeta.Len[i32, ComponentMeta]())
                    var a_comps = collections.NewVector[ComponentSnapshot](arch.columnMeta.Len[i32, ComponentMeta]())

                    var col_iter = arch.columnMeta.Iter[i32, ComponentMeta]()
                    var col_opt = col_iter.Next[i32, ComponentMeta]()
                    var col_loop = true
                    while (col_loop) {
                        match col_opt {
                            Some(col_entry) => {
                                var cid = col_entry.key
                                var ser_opt = w.serializers.Get[i32, ComponentSerializer](#cid)
                                match ser_opt {
                                    Some(ser) => {
                                        a_mask.Push[str]("" + ser.Name())
                                        var col_any_opt = arch.columns.Get[i32, any](#cid)
                                        match col_any_opt {
                                            Some(col_any) => {
                                                var comp_data = ser.Serialize(col_any)
                                                a_comps.Push[ComponentSnapshot](ComponentSnapshot {
                                                        name: "" + ser.Name(),
                                                        data: comp_data
                                                    })
                                            }
                                            None => {}
                                        }
                                    }
                                    None => {}
                                }
                                col_opt = col_iter.Next[i32, ComponentMeta]()
                            }
                            None => {
                                col_loop = false
                            }
                        }
                    }

                    var a_ents = make(Entity[], arch.entities.size)
                    for i in 0..arch.entities.size {
                        a_ents[i] = Entity {
                            id: arch.entities.data[i].id
                        }
                    }

                    var a_mask_arr = make(str[], a_mask.size)
                    for i in 0..a_mask.size {
                        a_mask_arr[i] = "" + a_mask.data[i]
                    }

                    var a_comps_arr = make(ComponentSnapshot[], a_comps.size)
                    for i in 0..a_comps.size {
                        var comp = #a_comps.data[i]
                        a_comps_arr[i] = ComponentSnapshot {
                            name: "" + comp.name,
                            data: "" + comp.data
                        }
                    }

                    snap_archetypes.Push[ArchSnapshot](ArchSnapshot {
                            mask: a_mask_arr,
                            entities: a_ents,
                            components: a_comps_arr
                        })
                }
                entry_opt = iter.Next[i64, @Archetype]()
            }
            None => {
                loop = false
            }
        }
    }

    var snap_ents_arr = make(EntityRecordRef[], snap_entities.size)
    for i in 0..snap_entities.size {
        snap_ents_arr[i] = EntityRecordRef {
            id: snap_entities.data[i].id,
            gen: snap_entities.data[i].gen
        }
    }

    var snap_archs_arr = make(ArchSnapshot[], snap_archetypes.size)
    for i in 0..snap_archetypes.size {
        var a = snap_archetypes.GetMut[ArchSnapshot](i)
        snap_archs_arr[i] = ArchSnapshot {
            mask: @a.mask,
            entities: @a.entities,
            components: @a.components
        }
    }

    return alloc Snapshot {
        next_id: w.nextEntityID,
        entities: snap_ents_arr,
        archetypes: snap_archs_arr
    }
}

pub fn RestoreSnapshot(data: str, w_orig: &World) @World {
    var snap = serialize.FromJSON[Snapshot](data)

    var w = alloc World {
        nextEntityID: snap.next_id,
        entities: collections.NewVector[EntityRecord](snap.next_id + 128),
        freeEntities: collections.NewVector[i32](128),
        archetypes: collections.NewHashMap[i64, @Archetype](64),
        root: 0,
        observers: collections.NewHashMap[i32, @ObserverTrigger](16),
        eventMgr: NewEventManager(),
        serializers: w_orig.serializers,
        nameToCid: w_orig.nameToCid
    }

    var root_arch = NewArchetype(0, NewBitSet())
    var root_ref: &Archetype = root_arch
    var root_hash = root_ref.mask.Hash()
    w.root = root_hash
    var _inserted_root = w.archetypes.Insert[i64, @Archetype](root_hash, root_arch)

    for i in 0..snap.next_id + 128 {
        w.entities.Push[EntityRecord](EntityRecord {
                gen: 0,
                archetype: 0,
                row: 0
            })
    }

    for i in 0..len(snap.archetypes) {
        var aSnap = #snap.archetypes[i]

        var mask = NewBitSet()
        var metas = collections.NewHashMap[i32, ComponentMeta](16)
        var columns = collections.NewHashMap[i32, any](16)

        for j in 0..len(aSnap.components) {
            var cSnap = #aSnap.components[j]
            var cid = 0
            var cid_opt = w.nameToCid.Get[str, i32](#cSnap.name)
            match cid_opt {
                Some(c) => {
                    cid = c
                }
                None => {}
            }
            if (cid > 0) {
                mask.Set(cid)
                var ser_opt = w.serializers.Get[i32, ComponentSerializer](#cid)
                match ser_opt {
                    Some(ser) => {
                        var col_any = ser.Deserialize("" + cSnap.data)
                        var m = ser.Meta()
                        var _i1 = columns.Insert[i32, any](cid + 0, col_any)
                        var _i2 = metas.Insert[i32, ComponentMeta](cid + 0, @m)
                    }
                    None => {}
                }
            }
        }

        var hash = mask.Hash()
        var arch = NewArchetype(w.archetypes.Len[i64, @Archetype](), mask)
        var arch_ref: &Archetype = arch

        arch_ref.columns = @columns
        arch_ref.columnMeta = @metas

        for j in 0..len(aSnap.entities) {
            var e = aSnap.entities[j]
            arch_ref.entities.Push[Entity](e)
        }

        for j in 0..arch_ref.entities.size {
            var e = arch_ref.entities.data[j]
            var id = e.ID()
            w.entities.data[id] = EntityRecord {
                gen: e.Gen(),
                archetype: hash,
                row: j
            }
        }

        var _i_arch = w.archetypes.Insert[i64, @Archetype](hash, arch)
    }

    return @w
}
