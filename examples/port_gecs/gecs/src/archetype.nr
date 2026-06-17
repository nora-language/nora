package gecs

import "collections"

pub fn NewArchetype(id: i32, mask: @BitSet) @Archetype {
    var a = alloc Archetype {
        id: id,
        mask: mask,
        entities: collections.NewVector[Entity](64),
        columns: collections.NewHashMap[i32, any](16),
        columnMeta: collections.NewHashMap[i32, ComponentMeta](16),
        addEdges: collections.NewHashMap[i32, i64](8),
        removeEdges: collections.NewHashMap[i32, i64](8)
    }
    return @a
}

pub fn (a: &Archetype) Alloc(e: Entity) i32 {
    var row = a.entities.size
    var e_val = e
    a.entities.Push[Entity](e_val)
    return row
}

pub fn (a: &Archetype) Remove(row: i32) @RemoveResult {
    var last_idx = a.entities.size - 1
    var moved_entity = NewEntity(0, 0)
    var ok = false

    if (row != last_idx) {
        var last_entity: Entity = @a.entities.data[last_idx]
        moved_entity = Entity {
            id: last_entity.id
        }
        a.entities.data[row] = last_entity
        ok = true
    }
    a.entities.size = a.entities.size - 1

    var iter = a.columnMeta.Iter[i32, ComponentMeta]()
    var entry_opt = iter.Next[i32, ComponentMeta]()

    var loop = true
    while (loop) {
        match entry_opt {
            Some(entry) => {
                var cid = entry.key
                var m = entry.value

                var col_any_opt = a.columns.GetMut[i32, any](#cid)
                match col_any_opt {
                    Some(col_any) => {
                        m.SwapRemove(col_any, row)
                    }
                    None => {}
                }

                entry_opt = iter.Next[i32, ComponentMeta]()
            }
            None => {
                loop = false
            }
        }
    }

    return alloc RemoveResult {
        moved_entity: moved_entity,
        ok: ok
    }
}

pub fn (a: @Archetype) drop() {
    var iter = a.columnMeta.Iter[i32, ComponentMeta]()
    var entry_opt = iter.Next[i32, ComponentMeta]()

    var loop = true
    while (loop) {
        match entry_opt {
            Some(entry) => {
                var cid = entry.key
                var m = entry.value

                var col_any_opt = a.columns.GetMut[i32, any](#cid)
                match col_any_opt {
                    Some(col_any) => {
                        m.DropCol(col_any)
                    }
                    None => {}
                }

                entry_opt = iter.Next[i32, ComponentMeta]()
            }
            None => {
                loop = false
            }
        }
    }

    a.entities = none
    a.columns = none
    a.columnMeta = none
    a.addEdges = none
    a.removeEdges = none
    a.mask = none
}
