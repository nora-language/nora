package gecs

import "collections"
import "io"
import "json"
import "serialize"
import "string"

pub fn NewObserverTrigger() @ObserverTrigger {
    return alloc ObserverTrigger {
        adds: collections.NewVector[ObserverCallback](4),
        removes: collections.NewVector[ObserverCallback](4),
        sets: collections.NewVector[ObserverCallback](4)
    }
}

pub fn NewWorld() @World {
    var w = alloc World {
        nextEntityID: 1,
        entities: collections.NewVector[EntityRecord](1000),
        freeEntities: collections.NewVector[i32](128),
        archetypes: collections.NewHashMap[i64, @Archetype](64),
        root: 0,
        observers: collections.NewHashMap[i32, @ObserverTrigger](16),
        eventMgr: NewEventManager(),
        serializers: collections.NewHashMap[i32, ComponentSerializer](16),
        nameToCid: collections.NewHashMap[str, i32](16)
    }
    var root_arch = NewArchetype(0, NewBitSet())
    var root_ref: &Archetype = root_arch
    var root_hash = root_ref.mask.Hash()
    w.root = root_hash
    var _inserted_root = w.archetypes.Insert[i64, @Archetype](root_hash, root_arch)
    return w
}

pub type ComponentSerializerImpl[T] = struct {
    name: str,
    cid: i32
}

pub fn (_s: &ComponentSerializerImpl[T]) Name[T]() str {
    return "" + _s.name
}

pub fn (_s: &ComponentSerializerImpl[T]) Serialize[T](col_any: &any) str {
    var vec: &collections.Vector[T] = collections.Vector[T](col_any)
    var parts = collections.NewVector[str](vec.size)
    for i in 0..vec.size {
        parts.Push(serialize.ToJSON[T](#vec.data[i]))
    }
    var inner = string.Join(parts, ", ")
    return "[${inner}]"
}

pub fn (_s: &ComponentSerializerImpl[T]) Deserialize[T](json_str: str) any {
    var vec = collections.NewVector[T](64)
    var parse_res = json.Parse(json_str)
    if (parse_res.IsErr()) {
        var res_err: any = @vec
        return res_err
    }
    var j_val = parse_res.Unwrap()
    var arr = j_val.GetArray()
    for i in 0..arr.Len() {
        var item = arr.Get(i)
        var item_str = json.Stringify(#item)
        var val = serialize.FromJSON[T](item_str)
        vec.Push[T](val)
    }
    var res_ok: any = @vec
    return res_ok
}

pub fn (_s: &ComponentSerializerImpl[T]) Meta[T]() ComponentMeta {
    var impl = alloc ComponentMetaImpl[T] {
        cid: _s.cid
    }
    var iface: ComponentMeta = impl
    return iface
}

pub fn RegisterComponent[T](w: &World, cid: i32, name: str) {
    var impl = alloc ComponentSerializerImpl[T] {
        name: "" + name,
        cid: cid
    }
    var ser: ComponentSerializer = impl
    var _inserted = w.serializers.Insert[i32, ComponentSerializer](cid + 0, @ser)
    var _inserted_name = w.nameToCid.Insert[str, i32](name, cid + 0)
}

pub type UnserializableComponentImpl[T] = struct {
    name: str,
    cid: i32
}

pub fn (_s: &UnserializableComponentImpl[T]) Name() str {
    return "" + _s.name
}

pub fn (_s: &UnserializableComponentImpl[T]) Serialize(col_any: &any) str {
    return "[]"
}

pub fn (_s: &UnserializableComponentImpl[T]) Deserialize(json_str: str) any {
    var vec = collections.NewVector[T](64)
    var res_ok: any = @vec
    return res_ok
}

pub fn (_s: &UnserializableComponentImpl[T]) Meta() ComponentMeta {
    var impl = alloc ComponentMetaImpl[T] {
        cid: _s.cid
    }
    var iface: ComponentMeta = impl
    return iface
}

pub fn RegisterUnserializableComponent[T](w: &World, cid: i32, name: str) {
    var impl = alloc UnserializableComponentImpl[T] {
        name: "" + name,
        cid: cid
    }
    var ser: ComponentSerializer = impl
    var _inserted = w.serializers.Insert[i32, ComponentSerializer](cid + 0, @ser)
    var _inserted_name = w.nameToCid.Insert[str, i32](name, cid + 0)
}

pub type ChildrenSerializerImpl = struct {
    name: str,
    cid: i32
}

pub fn (_s: &ChildrenSerializerImpl) Name() str {
    return "" + _s.name
}

pub fn (_s: &ChildrenSerializerImpl) Serialize(col_any: &any) str {
    var vec: &collections.Vector[Children] = collections.Vector[Children](col_any)
    var parts = collections.NewVector[str](vec.size)
    for i in 0..vec.size {
        var child_struct = &vec.data[i]
        var ids_slice = make(i64[], child_struct.entities.size)
        for j in 0..child_struct.entities.size {
            ids_slice[j] = child_struct.entities.data[j].id
        }
        var snap = ChildrenSnapshot {
            ids: @ids_slice
        }
        parts.Push(serialize.ToJSON[ChildrenSnapshot](#snap))
    }
    var inner = string.Join(parts, ", ")
    return "[${inner}]"
}

pub fn (_s: &ChildrenSerializerImpl) Deserialize(json_str: str) any {
    var vec = collections.NewVector[Children](64)
    var parse_res = json.Parse(json_str)
    if (parse_res.IsErr()) {
        var res_err: any = @vec
        return res_err
    }
    var j_val = parse_res.Unwrap()
    var arr = j_val.GetArray()
    for i in 0..arr.Len() {
        var item = arr.Get(i)
        var item_str = json.Stringify(#item)
        var snap = serialize.FromJSON[ChildrenSnapshot](item_str)
        var ent_vec = collections.NewVector[Entity](len(#snap.ids))
        for j in 0..len(#snap.ids) {
            ent_vec.Push[Entity](Entity {
                    id: (#snap.ids)[j]
                })
        }
        vec.Push[Children](Children {
                entities: ent_vec
            })
    }
    var res_ok: any = @vec
    return res_ok
}

pub fn (_s: &ChildrenSerializerImpl) Meta() ComponentMeta {
    var impl = alloc ComponentMetaImpl[Children] {
        cid: _s.cid
    }
    var iface: ComponentMeta = impl
    return iface
}

pub fn RegisterChildrenComponent(w: &World) {
    var impl = alloc ChildrenSerializerImpl {
        name: "children",
        cid: 4
    }
    var ser: ComponentSerializer = impl
    var _inserted = w.serializers.Insert[i32, ComponentSerializer](4, @ser)
    var _inserted_name = w.nameToCid.Insert[str, i32]("children", 4)
}

pub fn (w: &World) Spawn() Entity {
    var id = 0
    var gen = 0

    if (w.freeEntities.size > 0) {
        id = w.freeEntities.Pop[i32]()
        var record = @w.entities.data[id]
        gen = record.gen + 1
    } else {
        id = w.nextEntityID
        w.nextEntityID = w.nextEntityID + 1
        gen = 1

        while (w.entities.size <= id) {
            w.entities.Push[EntityRecord](EntityRecord {
                    gen: 0,
                    archetype: 0,
                    row: 0
                })
        }
    }

    var root_hash = w.root
    var root_opt = w.archetypes.GetMut[i64, @Archetype](#root_hash)
    match root_opt {
        Some(root_arch) => {
            w.entities.data[id] = EntityRecord {
                gen: gen,
                archetype: root_hash,
                row: root_arch.Alloc(NewEntity(id, gen))
            }
        }
        None => {
            panic("Root archetype not found")
        }
    }

    return NewEntity(id, gen)
}

pub fn (w: &World) reserve() Entity {
    var id = 0
    var gen = 0

    if (w.freeEntities.size > 0) {
        id = w.freeEntities.Pop[i32]()
        var record = @w.entities.data[id]
        gen = record.gen + 1
    } else {
        id = w.nextEntityID
        w.nextEntityID = w.nextEntityID + 1
        gen = 1

        while (w.entities.size <= id) {
            w.entities.Push[EntityRecord](EntityRecord {
                    gen: 0,
                    archetype: 0,
                    row: 0
                })
        }
    }

    w.entities.data[id] = EntityRecord {
        gen: gen,
        archetype: 0,
        row: 0
    }
    return NewEntity(id, gen)
}

pub fn (w: &World) spawnReserved(e: Entity) {
    var id = e.ID()
    var root_hash = w.root
    var root_opt = w.archetypes.GetMut[i64, @Archetype](#root_hash)
    match root_opt {
        Some(root_arch) => {
            w.entities.data[id].archetype = root_hash
            w.entities.data[id].row = root_arch.Alloc(e)
        }
        None => {
            panic("Root archetype not found")
        }
    }
}

pub fn (w: &World) Despawn(e: Entity) {
    var id = e.ID()
    if (!w.Alive(e)) {
        return
    }
    var record = w.entities.data[id]
    var arch_hash = record.archetype
    var arch_opt = w.archetypes.GetMut[i64, @Archetype](#arch_hash)
    match arch_opt {
        Some(arch) => {
            var remove_res = arch.Remove(record.row)
            if (remove_res.ok) {
                var moved_id = remove_res.moved_entity.ID()
                w.entities.data[moved_id].row = record.row
            }
        }
        None => {}
    }

    w.entities.data[id].archetype = 0
    w.freeEntities.Push[i32](id)
}

pub fn (w: #World) Alive(e: Entity) bool {
    var id = e.ID()
    if (id >= w.entities.size) {
        return false
    }
    var record = #w.entities.data[id]
    return record.gen == e.Gen() && record.archetype != 0
}

pub fn (w: &World) GetOrCreateNextArchetype(src_hash: i64, cid: i32, isAdd: bool) i64 {
    var next_hash: i64 = 0
    var src_opt = w.archetypes.GetMut[i64, @Archetype](#src_hash)
    match src_opt {
        Some(src) => {
            if (isAdd) {
                var opt = src.addEdges.Get[i32, i64](#cid)
                match opt {
                    Some(h) => {
                        next_hash = h
                    }
                    None => {}
                }
            } else {
                var opt = src.removeEdges.Get[i32, i64](#cid)
                match opt {
                    Some(h) => {
                        next_hash = h
                    }
                    None => {}
                }
            }
        }
        None => {
            panic("Source archetype not found")
        }
    }

    if (next_hash != 0) {
        return next_hash
    }

    var newMask = NewBitSet()
    var src_opt2 = w.archetypes.Get[i64, @Archetype](#src_hash)
    match src_opt2 {
        Some(src) => {
            newMask = src.mask.Clone()
        }
        None => {}
    }
    if (isAdd) {
        newMask.Set(cid + 0)
    } else {
        newMask.Clear(cid + 0)
    }
    var hash = newMask.Hash()

    var dest_opt = w.archetypes.Get[i64, @Archetype](#hash)
    var dest_exists = false
    match dest_opt {
        Some(d) => {
            dest_exists = true
        }
        None => {}
    }

    if (!dest_exists) {
        var new_arch = NewArchetype(w.archetypes.Len[i64, @Archetype](), newMask.Clone())
        var new_arch_ref: &Archetype = new_arch

        var src_opt3 = w.archetypes.Get[i64, @Archetype](#src_hash)
        match src_opt3 {
            Some(src) => {
                var iter = src.columnMeta.Iter[i32, ComponentMeta]()
                var entry_opt = iter.Next[i32, ComponentMeta]()

                var loop = true
                while (loop) {
                    match entry_opt {
                        Some(entry) => {
                            var k = entry.key
                            if (newMask.Has(k + 0)) {
                                var m = entry.value
                                var init_val = m.Init(64)
                                var ser_opt = w.serializers.Get[i32, ComponentSerializer](#k)
                                match ser_opt {
                                    Some(ser) => {
                                        var m_new = ser.Meta()
                                        var _inserted1 = new_arch_ref.columnMeta.Insert[i32, ComponentMeta](k + 0, m_new)
                                        var _inserted2 = new_arch_ref.columns.Insert[i32, any](k + 0, init_val)
                                    }
                                    None => {
                                        panic("Component serializer not found for k")
                                    }
                                }
                            }
                            entry_opt = iter.Next[i32, ComponentMeta]()
                        }
                        None => {
                            loop = false
                        }
                    }
                }
            }
            None => {}
        }

        if (isAdd) {
            var ser_opt = w.serializers.Get[i32, ComponentSerializer](#cid)
            match ser_opt {
                Some(ser) => {
                    var m = ser.Meta()
                    var init_val = m.Init(64)
                    var _inserted3 = new_arch_ref.columnMeta.Insert[i32, ComponentMeta](cid + 0, m)
                    var _inserted4 = new_arch_ref.columns.Insert[i32, any](cid + 0, init_val)
                }
                None => {
                    panic("Component serializer not found")
                }
            }
        }

        var hash_to_insert = hash + 0
        var _inserted_arch = w.archetypes.Insert[i64, @Archetype](hash_to_insert, new_arch)
    }

    var hash_val = hash + 0
    var src_opt4 = w.archetypes.GetMut[i64, @Archetype](#src_hash)
    var dest_opt2 = w.archetypes.GetMut[i64, @Archetype](#hash_val)

    match src_opt4 {
        Some(src) => {
            match dest_opt2 {
                Some(dest_arch) => {
                    var hash_copy1 = hash + 0
                    var hash_copy2 = hash + 0
                    if (isAdd) {
                        var _inserted5 = src.addEdges.Insert[i32, i64](cid + 0, hash_copy1)
                        var _inserted6 = dest_arch.removeEdges.Insert[i32, i64](cid + 0, src_hash)
                    } else {
                        var _inserted7 = src.removeEdges.Insert[i32, i64](cid + 0, hash_copy2)
                        var _inserted8 = dest_arch.addEdges.Insert[i32, i64](cid + 0, src_hash)
                    }
                }
                None => {}
            }
        }
        None => {}
    }

    var hash_final = hash + 0
    return hash_final
}

pub fn (w: &World) RegisterObserverAdd(cid: i32, f: fn(#Entity, &any)) {
    var obs_trigger: &@ObserverTrigger = none
    var obs_opt = w.observers.GetMut[i32, @ObserverTrigger](#cid)
    match obs_opt {
        Some(p) => {
            obs_trigger = p
        }
        None => {
            var trigger = NewObserverTrigger()
            var _inserted = w.observers.Insert[i32, @ObserverTrigger](cid + 0, trigger)
            var obs_opt2 = w.observers.GetMut[i32, @ObserverTrigger](#cid)
            match obs_opt2 {
                Some(p) => {
                    obs_trigger = p
                }
                None => {}
            }
        }
    }
    obs_trigger.adds.Push[ObserverCallback](ObserverCallback {
            f: f
        })
}

pub fn (w: &World) RegisterObserverSet(cid: i32, f: fn(#Entity, &any)) {
    var obs_trigger: &@ObserverTrigger = none
    var obs_opt = w.observers.GetMut[i32, @ObserverTrigger](#cid)
    match obs_opt {
        Some(p) => {
            obs_trigger = p
        }
        None => {
            var trigger = NewObserverTrigger()
            var _inserted = w.observers.Insert[i32, @ObserverTrigger](cid + 0, trigger)
            var obs_opt2 = w.observers.GetMut[i32, @ObserverTrigger](#cid)
            match obs_opt2 {
                Some(p) => {
                    obs_trigger = p
                }
                None => {}
            }
        }
    }
    obs_trigger.sets.Push[ObserverCallback](ObserverCallback {
            f: f
        })
}

pub fn (w: &World) RegisterObserverRemove(cid: i32, f: fn(#Entity, &any)) {
    var obs_trigger: &@ObserverTrigger = none
    var obs_opt = w.observers.GetMut[i32, @ObserverTrigger](#cid)
    match obs_opt {
        Some(p) => {
            obs_trigger = p
        }
        None => {
            var trigger = NewObserverTrigger()
            var _inserted = w.observers.Insert[i32, @ObserverTrigger](cid + 0, trigger)
            var obs_opt2 = w.observers.GetMut[i32, @ObserverTrigger](#cid)
            match obs_opt2 {
                Some(p) => {
                    obs_trigger = p
                }
                None => {}
            }
        }
    }
    obs_trigger.removes.Push[ObserverCallback](ObserverCallback {
            f: f
        })
}

pub fn (w: &World) triggerObserversAdd(e: Entity, cid: i32, val_any: &any) {
    var obs_opt = w.observers.Get[i32, @ObserverTrigger](#cid)
    match obs_opt {
        Some(trigger) => {
            for i in 0..trigger.adds.size {
                var cb = trigger.adds.Get[ObserverCallback](i)
                cb.f(Entity {
                        id: e.id
                    }, val_any)
            }
        }
        None => {}
    }
}

pub fn (w: &World) triggerObserversSet(e: Entity, cid: i32, val_any: &any) {
    var obs_opt = w.observers.Get[i32, @ObserverTrigger](#cid)
    match obs_opt {
        Some(trigger) => {
            for i in 0..trigger.sets.size {
                var cb = trigger.sets.Get[ObserverCallback](i)
                cb.f(Entity {
                        id: e.id
                    }, val_any)
            }
        }
        None => {}
    }
}

pub fn (w: &World) AddRaw(e: Entity, cid: i32, comp_any: any) {
    io.PrintLn("DEBUG AddRaw start")
    var id = e.ID()
    var record = w.entities.data[id]
    var r_row = record.row + 0
    var src_hash = record.archetype

    var src_opt = w.archetypes.GetMut[i64, @Archetype](#src_hash)
    match src_opt {
        Some(srcArch) => {
            if (srcArch.mask.Has(cid)) {
                io.PrintLn("DEBUG AddRaw has cid, setting")
                var col_any_opt = srcArch.columns.GetMut[i32, any](#cid)
                var meta_opt = srcArch.columnMeta.Get[i32, ComponentMeta](#cid)
                match col_any_opt {
                    Some(col_any) => {
                        match meta_opt {
                            Some(m) => {
                                m.Set(col_any, r_row, &comp_any)
                                w.triggerObserversSet(e, cid, &comp_any)
                            }
                            None => {}
                        }
                    }
                    None => {}
                }
                return
            }

            io.PrintLn("DEBUG AddRaw before GetOrCreateNextArchetype")
            var dest_hash = w.GetOrCreateNextArchetype(src_hash, cid, true)
            io.PrintLn("DEBUG AddRaw after GetOrCreateNextArchetype")
            var dest_opt = w.archetypes.GetMut[i64, @Archetype](#dest_hash)
            match dest_opt {
                Some(destArch) => {
                    io.PrintLn("DEBUG AddRaw before Alloc")
                    var destRow = destArch.Alloc(Entity {
                            id: e.id
                        })
                    io.PrintLn("DEBUG AddRaw after Alloc")

                    var iter = srcArch.columnMeta.Iter[i32, ComponentMeta]()
                    var entry_opt = iter.Next[i32, ComponentMeta]()

                    var loop = true
                    while (loop) {
                        match entry_opt {
                            Some(entry) => {
                                var k = entry.key
                                io.Print("DEBUG AddRaw move loop key = ")
                                io.PrintLn(io.FormatInt(k))
                                if (k + 0 != cid) {
                                    var m = entry.value

                                    var src_col_opt = srcArch.columns.GetMut[i32, any](#k)
                                    var dest_col_opt = destArch.columns.GetMut[i32, any](#k)

                                    match src_col_opt {
                                        Some(src_col) => {
                                            io.PrintLn("DEBUG AddRaw src_col found")
                                            match dest_col_opt {
                                                Some(dest_col) => {
                                                    io.PrintLn("DEBUG AddRaw dest_col found, moving")
                                                    m.MoveCol(src_col, r_row, dest_col)
                                                    io.PrintLn("DEBUG AddRaw move done")
                                                }
                                                None => {
                                                    io.PrintLn("DEBUG AddRaw dest_col NOT found")
                                                }
                                            }
                                        }
                                        None => {
                                            io.PrintLn("DEBUG AddRaw src_col NOT found")
                                        }
                                    }
                                }
                                entry_opt = iter.Next[i32, ComponentMeta]()
                            }
                            None => {
                                loop = false
                            }
                        }
                    }

                    io.PrintLn("DEBUG AddRaw after move loop")
                    var dest_col_opt = destArch.columns.GetMut[i32, any](#cid)
                    var meta_opt = destArch.columnMeta.Get[i32, ComponentMeta](#cid)
                    match dest_col_opt {
                        Some(dest_col) => {
                            match meta_opt {
                                Some(m) => {
                                    m.Add(dest_col, &comp_any)
                                }
                                None => {}
                            }
                        }
                        None => {}
                    }

                    io.PrintLn("DEBUG AddRaw after add_fn")
                    var remove_res = srcArch.Remove(r_row)
                    io.PrintLn("DEBUG AddRaw after srcArch.Remove")
                    if (remove_res.ok) {
                        var moved_id = remove_res.moved_entity.ID()
                        w.entities.data[moved_id].row = r_row
                    }

                    w.entities.data[id].archetype = dest_hash
                    w.entities.data[id].row = destRow

                    io.PrintLn("DEBUG AddRaw before triggerObserversAdd")
                    w.triggerObserversAdd(Entity {
                            id: e.id
                        }, cid, &comp_any)
                    io.PrintLn("DEBUG AddRaw done")
                }
                None => {}
            }
        }
        None => {}
    }
}

pub fn Add[T](w: &World, e: Entity, cid: i32, comp: T) {
    var val = comp
    var val_any: any = @val
    w.AddRaw(e, cid, val_any)
}

pub fn Remove[T](w: &World, e: Entity, cid: i32) {
    var id = e.ID()
    if (!w.Alive((Entity {
            id: e.id
        }))) {
        return
    }

    var record = w.entities.data[id]
    var r_row = record.row + 0
    var src_hash = record.archetype

    var src_opt = w.archetypes.GetMut[i64, @Archetype](#src_hash)
    match src_opt {
        Some(srcArch) => {
            if (!srcArch.mask.Has(cid)) {
                return
            }

            var obs_opt = w.observers.Get[i32, @ObserverTrigger](#cid)
            match obs_opt {
                Some(trigger) => {
                    if (trigger.removes.size > 0) {
                        var col_any_opt = srcArch.columns.GetMut[i32, any](#cid)
                        match col_any_opt {
                            Some(col_any) => {
                                var vec: &collections.Vector[T] = collections.Vector[T](col_any)
                                var val_ref = &vec.data[r_row]
                                var val_any: any = #val_ref
                                for i in 0..trigger.removes.size {
                                    var cb = trigger.removes.Get[ObserverCallback](i)
                                    cb.f(Entity {
                                            id: e.id
                                        }, &val_any)
                                }
                            }
                            None => {}
                        }
                    }
                }
                None => {}
            }

            var dest_hash = w.GetOrCreateNextArchetype(src_hash, cid, false)
            var dest_opt = w.archetypes.GetMut[i64, @Archetype](#dest_hash)
            match dest_opt {
                Some(destArch) => {
                    var destRow = destArch.Alloc(Entity {
                            id: e.id
                        })

                    var iter = srcArch.columnMeta.Iter[i32, ComponentMeta]()
                    var entry_opt = iter.Next[i32, ComponentMeta]()

                    var loop = true
                    while (loop) {
                        match entry_opt {
                            Some(entry) => {
                                var k = entry.key
                                if (k + 0 != cid) {
                                    var m = entry.value

                                    var src_col_opt = srcArch.columns.GetMut[i32, any](#k)
                                    var dest_col_opt = destArch.columns.GetMut[i32, any](#k)

                                    match src_col_opt {
                                        Some(src_col) => {
                                            match dest_col_opt {
                                                Some(dest_col) => {
                                                    m.MoveCol(src_col, r_row, dest_col)
                                                }
                                                None => {}
                                            }
                                        }
                                        None => {}
                                    }
                                }
                                entry_opt = iter.Next[i32, ComponentMeta]()
                            }
                            None => {
                                loop = false
                            }
                        }
                    }

                    var remove_res = srcArch.Remove(r_row)
                    if (remove_res.ok) {
                        var moved_id = remove_res.moved_entity.ID()
                        w.entities.data[moved_id].row = r_row
                    }

                    w.entities.data[id].archetype = dest_hash
                    w.entities.data[id].row = destRow
                }
                None => {}
            }
        }
        None => {}
    }
}

pub fn Get[T](w: #World, e: Entity, cid: i32) T {
    var id = e.ID()
    var record = w.entities.data[id]
    var arch_hash = record.archetype
    var arch_opt = w.archetypes.Get[i64, @Archetype](#arch_hash)
    match arch_opt {
        Some(arch) => {
            var r_row = record.row + 0
            var col_any_opt = arch.columns.Get[i32, any](#cid)
            match col_any_opt {
                Some(col_any) => {
                    var vec: &collections.Vector[T] = collections.Vector[T](col_any)
                    var val = vec.data[r_row]
                    return val
                }
                None => {
                    panic("Component column not found")
                }
            }
        }
        None => {
            panic("Archetype not found")
        }
    }
}

pub fn GetMut[T](w: &World, e: Entity, cid: i32) &T {
    var id = e.ID()
    var record = w.entities.data[id]
    var arch_hash = record.archetype
    var arch_opt = w.archetypes.GetMut[i64, @Archetype](#arch_hash)
    match arch_opt {
        Some(arch) => {
            var r_row = record.row + 0
            var col_any_opt = arch.columns.GetMut[i32, any](#cid)
            match col_any_opt {
                Some(col_any) => {
                    var vec: &collections.Vector[T] = collections.Vector[T](col_any)
                    return &vec.data[r_row]
                }
                None => {
                    panic("Component column not found")
                }
            }
        }
        None => {
            panic("Archetype not found")
        }
    }
}

pub fn Has(w: #World, e: Entity, cid: i32) bool {
    var id = e.ID()
    var record = w.entities.data[id]
    if (record.archetype == 0) {
        return false
    }
    var arch_hash = record.archetype
    var arch_opt = w.archetypes.Get[i64, @Archetype](#arch_hash)
    match arch_opt {
        Some(arch) => {
            return arch.mask.Has(cid)
        }
        None => {
            return false
        }
    }
}
