package gecs

import "collections"

pub fn NewEntity(index: i32, gen: i32) Entity {
    var id_i64 = i64(index)
    var gen_i64 = i64(gen)
    return Entity {
        id: (gen_i64 << 32) | (id_i64 & 4294967295)
    }
}

pub fn (self: Entity) ID() i32 {
    return i32(self.id & 4294967295)
}

pub fn (self: Entity) Gen() i32 {
    return i32(self.id >> 32)
}

pub fn (self: Entity) Equals(other: Entity) bool {
    return self.id == other.id
}

pub fn InitColumn[T](cap: i32) any {
    var vec = collections.NewVector[T](cap)
    var res: any = @vec
    return res
}

pub fn SwapRemoveColumn[T](col_any: &any, row: i32) void {
    var vec: &collections.Vector[T] = collections.Vector[T](col_any)
    if (vec.size > 0) {
        var last_idx = vec.size - 1
        if (row != last_idx) {
            var last_val = vec.data[last_idx]
            vec.data[row] = @last_val
        }
        vec.size = vec.size - 1
    }
}

pub fn MoveColumn[T](src_any: &any, src_row: i32, dest_any: &any) {
    var src_vec: &collections.Vector[T] = collections.Vector[T](src_any)
    var dest_vec: &collections.Vector[T] = collections.Vector[T](dest_any)

    var val: T = @src_vec.data[src_row]
    dest_vec.Push[T](val)
}

pub fn DropColumn[T](_col_any: &any) {
    // No-op. The map containing the any box will automatically
    // clean up and drop the Vector[T] when it is destroyed.
}

pub fn AddRawColumn[T](col_any: &any, val_any: &any) void {
    var vec: &collections.Vector[T] = collections.Vector[T](col_any)
    var val: &T = T(val_any)
    var temp: T = @val
    vec.Push[T](temp)
}

pub fn SetRawColumn[T](col_any: &any, row: i32, val_any: &any) void {
    var vec: &collections.Vector[T] = collections.Vector[T](col_any)
    var val: &T = T(val_any)
    var temp: T = @val

    if (row >= vec.capacity) {
        var new_cap = vec.capacity * 2
        if (new_cap <= row) {
            new_cap = row + 1
        }
        var new_data = alloc T[new_cap]
        for i in 0..vec.size {
            new_data[i] = @vec.data[i]
        }
        vec.data = @new_data
        vec.capacity = new_cap
    }

    if (row >= vec.size) {
        vec.size = row + 1
    }

    vec.data[row] = temp
}

pub type ComponentMetaImpl[T] = struct {
    cid: i32
}

pub fn (_m: &ComponentMetaImpl[T]) Init[T](cap: i32) any {
    return InitColumn[T](cap)
}

pub fn (_m: &ComponentMetaImpl[T]) SwapRemove[T](col_any: &any, row: i32) void {
    SwapRemoveColumn[T](col_any, row)
}

pub fn (_m: &ComponentMetaImpl[T]) MoveCol[T](src_any: &any, src_row: i32, dest_any: &any) void {
    MoveColumn[T](src_any, src_row, dest_any)
}

pub fn (_m: &ComponentMetaImpl[T]) DropCol[T](col_any: &any) void {
    DropColumn[T](col_any)
}

pub fn (_m: &ComponentMetaImpl[T]) Add[T](col_any: &any, val_any: &any) void {
    AddRawColumn[T](col_any, val_any)
}

pub fn (_m: &ComponentMetaImpl[T]) Set[T](col_any: &any, row: i32, val_any: &any) void {
    SetRawColumn[T](col_any, row, val_any)
}
