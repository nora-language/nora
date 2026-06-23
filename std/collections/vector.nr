package collections

import "io"

[core_intrinsic("Collection")]
pub type Vector[T] = struct {
    data: @T[],
    size: i32,
    capacity: i32
}

pub type VectorIterator[T] = struct {
    vec: #Vector[T],
    idx: i32
}

pub fn NewVector[T](cap: i32) @Vector[T] {
    var c = cap
    if (c < 1) {
        c = 1
    }
    return alloc Vector[T] {
        data: alloc T[c],
        size: 0,
        capacity: c
    }
}

pub fn (v: #Vector[T]) Len[T]() i32 {
    return v.size
}

[unsafe]
pub fn (v: &Vector[T]) Push[T](val: T) {
    if (v.size == v.capacity) {
        var new_cap = v.capacity * 2
        var new_data = alloc T[new_cap]

        var i = 0
        while i < v.size {
            new_data.unchecked_set(i, @v.data.unchecked_get(i))
            i++
        }

        v.data = @new_data
        v.capacity = new_cap
    }

    v.data.unchecked_set(v.size, @val)
    v.size = v.size + 1
}

pub fn (v: #Vector[T]) Get[T](index: i32) #T {
    if (index < 0 || index >= v.size) {
        io.PrintLn("Error: Index ${index} out of bounds (size ${v.size})")
    }
    return v.data.unchecked_get(index)
}

[unsafe]
pub fn (v: &Vector[T]) GetMut[T](index: i32) &T {
    if (index < 0 || index >= v.size) {
        io.PrintLn("Error: Index ${index} out of bounds (size ${v.size})")
    }
    return &v.data.unchecked_get(index)
}


[unsafe]
pub fn (v: &Vector[T]) Pop[T]() T {
    if (v.size == 0) {
        io.PrintLn("Error: Pop from empty vector")
    }
    v.size = v.size - 1
    return v.data[v.size]
}

[unsafe]
pub fn (v: &Vector[T]) Set[T](index: i32, val: T) {
    if (index >= 0 && index < v.size) {
        v.data[index] = val
    }
}

[unsafe]
pub fn (v: &Vector[T]) Remove[T](index: i32) T {
    if (index < 0 || index >= v.size) {
        io.PrintLn("Error: Index ${index} out of bounds (size ${v.size})")
    }
    var val: T = @v.data.unchecked_get(index)
    var i = index
    while i < v.size - 1 {
        v.data.unchecked_set(i, @v.data.unchecked_get(i + 1))
        i = i + 1
    }
    v.size = v.size - 1
    return val
}

[unsafe]
pub fn (v: &Vector[T]) Insert[T](index: i32, val: T) {
    if (index < 0 || index > v.size) {
        io.PrintLn("Error: Index ${index} out of bounds (size ${v.size})")
    }
    if (v.size == v.capacity) {
        var new_cap = v.capacity * 2
        var new_data = alloc T[new_cap]
        var i = 0
        while i < v.size {
            new_data.unchecked_set(i, @v.data.unchecked_get(i))
            i = i + 1
        }
        v.data = @new_data
        v.capacity = new_cap
    }
    var i = v.size
    while i > index {
        v.data.unchecked_set(i, @v.data.unchecked_get(i - 1))
        i = i - 1
    }
    v.data.unchecked_set(index, val)
    v.size = v.size + 1
}

[unsafe]
pub fn (v: &Vector[T]) Clear[T]() {
    var new_data = alloc T[v.capacity]
    v.data = @new_data
    v.size = 0
}

[unsafe]
pub fn (v: &Vector[T]) Resize[T](new_size: i32) {
    if (new_size < 0) {
        io.PrintLn("Error: Resize size cannot be negative")
    }
    if (new_size < v.size) {
        var new_data = alloc T[new_size]
        var i = 0
        while i < new_size {
            new_data.unchecked_set(i, @v.data.unchecked_get(i))
            i = i + 1
        }
        v.data = @new_data
        v.capacity = new_size
        v.size = new_size
    } else if (new_size > v.size) {
        if (new_size > v.capacity) {
            var new_data = alloc T[new_size]
            var i = 0
            while i < v.size {
                new_data.unchecked_set(i, @v.data.unchecked_get(i))
                i = i + 1
            }
            v.data = @new_data
            v.capacity = new_size
        }
        v.size = new_size
    }
}

pub fn (v: #Vector[T]) ForEach[T](f: fn(#T)) {
    var i = 0
    while i < v.size {
        f(#v.data[i])
        i = i + 1
    }
}

pub fn (v: #Vector[T]) Filter[T](predicate: fn(#T) bool) @Vector[T] {
    var res = NewVector[T](v.size)
    var i = 0
    while i < v.size {
        if (predicate(#v.data.unchecked_get(i))) {
            res.Push(v.data.unchecked_get(i))
        }
        i = i + 1
    }
    return res
}

pub fn (v: #Vector[T]) Map[T, U](mapper: fn(#T) U) @Vector[U] {
    var res = NewVector[U](v.size)
    var i = 0
    while i < v.size {
        res.Push(mapper(#v.data.unchecked_get(i)))
        i = i + 1
    }
    return res
}

pub fn (self: &VectorIterator[T]) Next[T]() Option[#T] {
    if (self.idx >= self.vec.size) {
        return None[#T]
    }
    var val = #self.vec.data.unchecked_get(self.idx)
    self.idx = self.idx + 1
    return Some[#T](val)
}

pub fn (v: #Vector[T]) Iter[T]() VectorIterator[T] {
    return VectorIterator[T] {
        vec: v,
        idx: 0
    }
}

[unsafe]
fn run_par_map_worker[T](val: &T, f: #fn(&T)) {
    f(val)
}

[unsafe]
pub fn (v: &Vector[T]) ParMap[T](f: fn(&T)) {
    scope {
        var i = 0
        while i < v.size {
            var ref = v.GetMut[T](i)
            spawn run_par_map_worker[T](ref, #f)
            i = i + 1
        }
    }
}





