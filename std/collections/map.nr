package collections

import "io"
import "hash"

type MapEntry[K, V] = struct {
    key: K,
    value: V,
    occupied: bool,
    deleted: bool
}

[core_intrinsic("Collection")]
pub type HashMap[K: hash.Hashable, V] = struct {
    entries: @MapEntry[K, V][],
    size: i32,
    capacity: i32
}

pub type MapKeyValuePair[K, V] = struct {
    key: K,
    value: V
}

pub type MapIterator[K: hash.Hashable, V] = struct {
    map: #HashMap[K, V],
    idx: i32
}

pub fn NewHashMap[K: hash.Hashable, V](cap: i32) @HashMap[K, V] {
    var c = cap
    if (c < 2) {
        c = 2
    }
    return alloc HashMap[K, V] {
        entries: alloc MapEntry[K, V][c],
        size: 0,
        capacity: c
    }
}

pub fn (m: #HashMap[K, V]) Len[K: hash.Hashable, V]() i32 {
    return m.size
}

[unsafe]
pub fn (m: &HashMap[K, V]) rehash[K: hash.Hashable, V](new_cap: i32) {
    var old_cap = m.capacity
    var old_entries = @m.entries

    m.capacity = new_cap
    m.entries = alloc MapEntry[K, V][new_cap]
    m.size = 0

    var i = 0
    while i < old_cap {
        if (old_entries[i].occupied) {
            m.Insert(@old_entries[i].key, @old_entries[i].value)
        }
        i = i + 1
    }
}

[unsafe]
pub fn (m: &HashMap[K, V]) Insert[K: hash.Hashable, V](key: K, value: V) bool {
    // 70% load factor threshold
    if (m.size * 10 >= m.capacity * 7) {
        m.rehash[K, V](m.capacity * 2)
    }

    var h = key.Hash()
    if (h < 0) {
        h = 0 - h
    }
    var idx = h % m.capacity
    var first_deleted = -1

    var i = 0
    while i < m.capacity {
        var entry_idx = (idx + i) % m.capacity
        if (m.entries[entry_idx].occupied) {
            if (m.entries[entry_idx].key == key) {
                var _old_val = m.entries[entry_idx].value
                m.entries[entry_idx].value = value
                return false
            }
        } else {
            if (m.entries[entry_idx].deleted) {
                if (first_deleted == -1) {
                    first_deleted = entry_idx
                }
            } else {
                var insert_idx = entry_idx
                if (first_deleted != -1) {
                    insert_idx = first_deleted
                }
                m.entries[insert_idx].key = key
                m.entries[insert_idx].value = value
                m.entries[insert_idx].occupied = true
                m.entries[insert_idx].deleted = false
                m.size = m.size + 1
                return true
            }
        }
        i = i + 1
    }

    if (first_deleted != -1) {
        m.entries[first_deleted].key = key
        m.entries[first_deleted].value = value
        m.entries[first_deleted].occupied = true
        m.entries[first_deleted].deleted = false
        m.size = m.size + 1
        return true
    }
    return false
}

pub fn (m: #HashMap[K, V]) Get[K: hash.Hashable, V](key: #K) Option[#V] {
    var h = key.Hash()
    if (h < 0) {
        h = 0 - h
    }
    var idx = h % m.capacity

    var i = 0
    while i < m.capacity {
        var entry_idx = (idx + i) % m.capacity
        if (m.entries[entry_idx].occupied) {
            if (m.entries[entry_idx].key == key) {
                return Some[#V](#m.entries[entry_idx].value)
            }
        } else {
            if (m.entries[entry_idx].deleted == false) {
                return None[#V]
            }
        }
        i = i + 1
    }
    return None[#V]
}

[unsafe]
pub fn (m: &HashMap[K, V]) GetMut[K: hash.Hashable, V](key: #K) Option[&V] {
    var h = key.Hash()
    if (h < 0) {
        h = 0 - h
    }
    var idx = h % m.capacity

    var i = 0
    while i < m.capacity {
        var entry_idx = (idx + i) % m.capacity
        if (m.entries[entry_idx].occupied) {
            if (m.entries[entry_idx].key == key) {
                return Some[&V](&m.entries[entry_idx].value)
            }
        } else {
            if (m.entries[entry_idx].deleted == false) {
                return None[&V]
            }
        }
        i = i + 1
    }
    return None[&V]
}

pub fn (m: #HashMap[K, V]) Contains[K: hash.Hashable, V](key: #K) bool {
    var h = key.Hash()
    if (h < 0) {
        h = 0 - h
    }
    var idx = h % m.capacity

    var i = 0
    while i < m.capacity {
        var entry_idx = (idx + i) % m.capacity
        if (m.entries[entry_idx].occupied) {
            if (m.entries[entry_idx].key == key) {
                return true
            }
        } else {
            if (m.entries[entry_idx].deleted == false) {
                return false
            }
        }
        i = i + 1
    }
    return false
}

[unsafe]
pub fn (m: &HashMap[K, V]) Remove[K: hash.Hashable, V](key: #K) Option[V] {
    var h = key.Hash()
    if (h < 0) {
        h = 0 - h
    }
    var idx = h % m.capacity

    var i = 0
    while i < m.capacity {
        var entry_idx = (idx + i) % m.capacity
        if (m.entries[entry_idx].occupied) {
            if (m.entries[entry_idx].key == key) {
                m.entries[entry_idx].occupied = false
                m.entries[entry_idx].deleted = true
                m.size = m.size - 1
                return Some[V](@m.entries[entry_idx].value)
            }
        } else {
            if (m.entries[entry_idx].deleted == false) {
                return None[V]
            }
        }
        i = i + 1
    }
    return None[V]
}

[unsafe]
pub fn (self: &MapIterator[K, V]) Next[K: hash.Hashable, V]() Option[MapKeyValuePair[#K, #V]] {
    while (self.idx < self.map.capacity) {
        var entry = #self.map.entries[self.idx]
        self.idx = self.idx + 1
        if (entry.occupied) {
            return Some(MapKeyValuePair[#K, #V] {
                key: #entry.key,
                value: #entry.value
            })
        }
    }
    return None[MapKeyValuePair[#K, #V]]
}

pub fn (m: #HashMap[K, V]) Iter[K: hash.Hashable, V]() MapIterator[K, V] {
    return MapIterator[K, V] {
        map: m,
        idx: 0
    }
}
