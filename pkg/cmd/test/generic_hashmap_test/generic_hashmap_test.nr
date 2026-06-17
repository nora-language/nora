package main

import "io"

extern fn strlen(s: #str) i32

type Entry[K, V] = struct {
    key: K,
    value: V,
    next: @Entry[K, V]
}

type HashMap[K, V] = struct {
    buckets: @(@Entry[K, V])[],
    capacity: i32,
    size: i32
}

fn new_hashmap[K, V](cap: i32) @HashMap[K, V] {
    var buckets = alloc (@Entry[K, V])[cap]
    var i = 0
    while i < cap {
        buckets[i] = none
        i++
    }
    return alloc HashMap[K, V] {
        buckets: @buckets,
        capacity: cap,
        size: 0
    }
}

fn hash_str(key: #str) i32 {
    var h = 0
    var i = 0
    var len_key = strlen(key)
    while (i < len_key && i < 100) {
        var c = key[i]
        h = (h * 31) + c
        i = i + 1
    }
    if (h < 0) {
        h = -h
    }
    return h
}

fn (self: &HashMap[K, V]) put[K, V](key: K, val: V, hash_fn: fn(#K) i32) {
    if (self.size > (self.capacity * 3 / 4)) {
        resize_hashmap[K, V](self, self.capacity * 2, hash_fn)
    }

    var h = hash_fn(#key)
    var idx = h % self.capacity
    var curr = &self.buckets[idx]

    while (curr != none) {
        if (curr.key == key) {
            curr.value = val
            return
        }
        curr = &curr.next
    }

    var new_entry = alloc Entry[K, V] {
        key: key,
        value: val,
        next: @self.buckets[idx]
    }
    self.buckets[idx] = @new_entry
    self.size = self.size + 1
}

fn (self: #HashMap[K, V]) get[K, V](key: #K, hash_fn: fn(#K) i32, default_val: V) V {
    var h = hash_fn(key)
    var idx = h % self.capacity
    var curr = #self.buckets[idx]
    while (curr != none) {
        if (curr.key == key) {
            return curr.value
        }
        curr = #curr.next
    }
    return default_val
}

fn (self: &HashMap[K, V]) remove[K, V](key: #K, hash_fn: fn(#K) i32) {
    var h = hash_fn(key)
    var idx = h % self.capacity
    var curr = &self.buckets[idx]
    if (curr == none) {
        return
    }

    if (curr.key == key) {
        var to_drop = @self.buckets[idx]
        self.buckets[idx] = @to_drop.next
        self.size = self.size - 1
        return
    }

    var prev = &self.buckets[idx]
    var node = &prev.next
    while (node != none) {
        if (node.key == key) {
            var to_drop = @prev.next
            prev.next = @to_drop.next
            self.size = self.size - 1
            return
        }
        prev = node
        node = &node.next
    }
}

fn resize_hashmap[K, V](self: &HashMap[K, V], new_cap: i32, hash_fn: fn(#K) i32) {
    io.PrintLn("Resizing HashMap: ${self.capacity} -> ${new_cap}")
    var old_buckets = @self.buckets
    var old_cap = self.capacity

    var new_buckets = alloc (@Entry[K, V])[new_cap]
    var i = 0
    while i < new_cap {
        new_buckets[i] = none
        i++
    }

    self.buckets = @new_buckets
    self.capacity = new_cap
    self.size = 0

    i = 0
    while i < old_cap {
        while (old_buckets[i] != none) {
            match @old_buckets[i] {
                Entry {
                    key: k,
                    value: v,
                    next: n
                } => {
                    var h = hash_fn(#k)
                    var idx = h % self.capacity

                    self.buckets[idx] = alloc Entry[K, V] {
                        key: @k,
                        value: @v,
                        next: @self.buckets[idx]
                    }
                    self.size = self.size + 1

                    // Update the current bucket to point to the next node
                    old_buckets[i] = @n
                }
            }
        }
        i++
    }
}

fn hash_i32(key: i32) i32 {
    if (key < 0) {
        return -key
    }
    return key
}

fn test_int_keys() {
    io.PrintLn("\nStep 5a: Testing HashMap[i32, str]")
    var m = new_hashmap[i32, str](8)

    m.put(1, "one", hash_i32)
    m.put(10, "ten", hash_i32)
    m.put(100, "hundred", hash_i32)

    var s1 = m.get(1, hash_i32, "unknown")
    var s2 = m.get(10, hash_i32, "unknown")
    var s3 = m.get(100, hash_i32, "unknown")
    var s4 = m.get(5, hash_i32, "unknown")

    io.PrintLn("1: ${s1}")
    io.PrintLn("10: ${s2}")
    io.PrintLn("100: ${s3}")
    io.PrintLn("5: ${s4}")

    if (s1 == "one" && s2 == "ten" && s3 == "hundred" &&
        s4 == "unknown") {
        io.PrintLn("Int keys test passed!")
    } else {
        io.PrintLn("Int keys test failed!")
    }
}

fn test_str_str_map() {
    io.PrintLn("\nStep 5b: Testing HashMap[str, str]")
    var m = new_hashmap[str, str](8)

    m.put("red", "#FF0000", hash_str)
    m.put("green", "#00FF00", hash_str)
    m.put("blue", "#0000FF", hash_str)

    var c1 = m.get("red", hash_str, "none")
    var c2 = m.get("blue", hash_str, "none")

    io.PrintLn("red: ${c1}")
    io.PrintLn("blue: ${c2}")

    if (c1 == "#FF0000" && c2 == "#0000FF") {
        io.PrintLn("Str-Str map test passed!")
    } else {
        io.PrintLn("Str-Str map test failed!")
    }
}

fn main() i32 {
    io.PrintLn("Step 4: Testing resize")
    var m = new_hashmap[str, i32](4)

    m.put("apple", 100, hash_str)
    m.put("banana", 200, hash_str)
    m.put("cherry", 300, hash_str)
    m.put("date", 400, hash_str) // Should trigger resize
    m.put("elderberry", 500, hash_str)

    io.PrintLn("Size after insertions: ${m.size}")
    io.PrintLn("Capacity after insertions: ${m.capacity}")

    var v1 = m.get("apple", hash_str, -1)
    var v2 = m.get("date", hash_str, -1)
    var v5 = m.get("elderberry", hash_str, -1)

    io.PrintLn("apple: ${v1}")
    io.PrintLn("date: ${v2}")
    io.PrintLn("elderberry: ${v5}")

    if (v1 == 100 && v2 == 400 && v5 == 500 && m.capacity > 4) {
        io.PrintLn("Step 4 passed!")
    } else {
        io.PrintLn("Step 4 failed!")
        return 1
    }

    test_int_keys()
    test_str_str_map()

    return 0
}
