package main

import "io"

extern fn strlen(s: #str) i32

type Entry = struct {
    key: str,
    value: i32,
    next: @Entry
}

type CustomMap = struct {
    buckets: @(@Entry)[],
    capacity: i32
}

fn create_map(cap: i32) @CustomMap {
    var buckets = alloc (@Entry)[cap]
    var i = 0
    while (i < cap) {
        buckets[i] = none
        i = i + 1
    }
    return alloc CustomMap {
        buckets: @buckets,
        capacity: cap
    }
}

fn (self: #CustomMap) hash(key: #str) i32 {
    var h = 0
    var i = 0
    var len_key = strlen(key)
    while (i < len_key && i < 10) {
        var c = key[i]
        h = (h * 31) + c
        i = i + 1
    }
    if (h < 0) {
        h = -h
    }
    return h % self.capacity
}

fn (self: &CustomMap) put(key: str, val: i32) {
    var idx = self.hash(key)
    var curr = &self.buckets[idx]

    while (curr != none) {
        if (curr.key == key) {
            curr.value = val
            return
        }
        curr = &curr.next
    }

    var new_entry = alloc Entry {
        key: key,
        value: val,
        next: @self.buckets[idx]
    }
    self.buckets[idx] = @new_entry
}

fn (self: #CustomMap) get(key: #str) i32 {
    var idx = self.hash(key)
    var curr = #self.buckets[idx]
    while (curr != none) {
        if (curr.key == key) {
            return curr.value
        }
        curr = #curr.next
    }
    return -1
}

fn (self: &CustomMap) remove(key: #str) {
    var idx = self.hash(key)
    var curr = &self.buckets[idx]
    if (curr == none) {
        return
    }

    if (curr.key == key) {
        var to_drop = @self.buckets[idx]
        self.buckets[idx] = @to_drop.next
        return
    }

    var prev = &self.buckets[idx]
    var node = &prev.next
    while (node != none) {
        if (node.key == key) {
            var to_drop = @prev.next
            prev.next = @to_drop.next
            return
        }
        prev = node
        node = &node.next
    }
}

fn main() i32 {
    io.PrintLn("Step 3a: CustomMap creation and put")
    var m = create_map(16)
    m.put("apple", 100)
    m.put("banana", 200)
    io.PrintLn("Put successful")

    io.PrintLn("Step 3b: CustomMap get")
    var v1 = m.get("apple")
    var v2 = m.get("banana")
    if (v1 == 100) {
        io.PrintLn("Get apple: OK")
    }
    if (v2 == 200) {
        io.PrintLn("Get banana: OK")
    }

    io.PrintLn("Step 3c: CustomMap remove")
    m.remove("apple")
    var v4 = m.get("apple")
    if (v4 == -1) {
        io.PrintLn("Remove apple: OK")
    }

    var v5 = m.get("banana")
    if (v5 == 200) {
        io.PrintLn("Banana still exists: OK")
    }

    return 0
}
