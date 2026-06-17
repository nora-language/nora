package gecs

pub fn NewBitSet() @BitSet {
    var cap = 2
    var words = alloc i64[cap]
    for i in 0..cap {
        words[i] = 0
    }
    return alloc BitSet {
        words: words,
        size: 0,
        capacity: cap
    }
}

pub fn (self: &BitSet) Set(id: i32) {
    var word_idx = id / 64
    if (word_idx >= self.capacity) {
        var new_cap = self.capacity * 2
        if (new_cap <= word_idx) {
            new_cap = word_idx + 1
        }
        var new_words = alloc i64[new_cap]
        for i in 0..new_cap {
            new_words[i] = 0
        }
        for i in 0..self.capacity {
            new_words[i] = self.words[i]
        }
        self.words = @new_words
        self.capacity = new_cap
    }

    var word = self.words[word_idx]
    var shift = i64(id % 64)
    var bitmask = i64(1) << shift
    word = word | bitmask
    self.words[word_idx] = word
}

pub fn (self: &BitSet) Clear(id: i32) {
    var word_idx = id / 64
    if (word_idx >= self.capacity) {
        return
    }

    var word = self.words[word_idx]
    var shift = i64(id % 64)
    var bitmask = i64(1) << shift
    var minus_one = i64(0) - i64(1)
    word = word & (minus_one ^ bitmask)
    self.words[word_idx] = word
}

pub fn (self: #BitSet) Has(id: i32) bool {
    var word_idx = id / 64
    if (word_idx >= self.capacity) {
        return false
    }
    var word = self.words[word_idx]
    var shift = i64(id % 64)
    var bitmask = i64(1) << shift
    return (word & bitmask) != 0
}

pub fn (self: #BitSet) Equals(other: #BitSet) bool {
    var max_cap = self.capacity
    if (other.capacity > max_cap) {
        max_cap = other.capacity
    }

    for i in 0..max_cap {
        var w1 = 0
        if (i < self.capacity) {
            w1 = self.words[i]
        }
        var w2 = 0
        if (i < other.capacity) {
            w2 = other.words[i]
        }
        if (w1 != w2) {
            return false
        }
    }
    return true
}

pub fn (self: #BitSet) Contains(other: #BitSet) bool {
    if (other.capacity > self.capacity) {
        for i in self.capacity..other.capacity {
            var w2 = other.words[i]
            if (w2 != 0) {
                return false
            }
        }
    }

    var min_cap = self.capacity
    if (other.capacity < min_cap) {
        min_cap = other.capacity
    }

    for i in 0..min_cap {
        var w1 = self.words[i]
        var w2 = other.words[i]
        if ((w1 & w2) != w2) {
            return false
        }
    }
    return true
}

pub fn (self: #BitSet) Intersects(other: #BitSet) bool {
    var min_cap = self.capacity
    if (other.capacity < min_cap) {
        min_cap = other.capacity
    }

    for i in 0..min_cap {
        var w1 = self.words[i]
        var w2 = other.words[i]
        if ((w1 & w2) != 0) {
            return true
        }
    }
    return false
}

pub fn (self: #BitSet) Clone() @BitSet {
    var new_words = alloc i64[self.capacity]
    for i in 0..self.capacity {
        new_words[i] = self.words[i]
    }
    return alloc BitSet {
        words: new_words,
        size: self.size,
        capacity: self.capacity
    }
}

pub fn (self: #BitSet) Hash() i64 {
    var hash = i64(0) - 4132072383898762747
    for i in 0..self.capacity {
        var w = self.words[i]
        hash = hash ^ w
        hash = hash * 1099511628211
    }
    return hash
}
